#!/usr/bin/env bash
set -euo pipefail

# XueHua Server build, runtime lifecycle, and release packaging manager.
# Designed to work both in a source checkout and inside an extracted release archive
# (where neither Go nor Mage is required to run start/stop/check).

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUTPUT_DIR="${OUTPUT_DIR:-_output}"
CONFIG_DIR="${CONFIG_DIR:-config}"
START_CONFIG="${START_CONFIG:-start-config.yml}"
VERSION="${VERSION:-dev}"
RELEASE="${RELEASE:-false}"
PLATFORMS="${PLATFORMS:-}"
JOBS="${JOBS:-4}"

case "$OUTPUT_DIR" in
    /*) OUTPUT_ABS="$OUTPUT_DIR" ;;
    *)  OUTPUT_ABS="$ROOT_DIR/$OUTPUT_DIR" ;;
esac

case "$CONFIG_DIR" in
    /*) CONFIG_ABS="$CONFIG_DIR" ;;
    *)  CONFIG_ABS="$ROOT_DIR/$CONFIG_DIR" ;;
esac

case "$START_CONFIG" in
    /*) START_CONFIG_ABS="$START_CONFIG" ;;
    *)  START_CONFIG_ABS="$ROOT_DIR/$START_CONFIG" ;;
esac

if [[ "${DEPLOYMENT_TYPE:-}" == "kubernetes" ]]; then
    CONFIG_ABS="/config"
fi

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;36m'
NC='\033[0m'

log_info()  { printf "${BLUE}[INFO]${NC} %s\n" "$*"; }
log_ok()    { printf "${GREEN}[OK]${NC} %s\n" "$*"; }
log_warn()  { printf "${YELLOW}[WARN]${NC} %s\n" "$*"; }
log_error() { printf "${RED}[ERROR]${NC} %s\n" "$*" >&2; }

host_os() {
    case "$(uname -s | tr '[:upper:]' '[:lower:]')" in
        darwin) echo "darwin" ;;
        linux)  echo "linux" ;;
        msys*|mingw*|cygwin*) echo "windows" ;;
        *) uname -s | tr '[:upper:]' '[:lower:]' ;;
    esac
}

host_arch() {
    case "$(uname -m)" in
        x86_64|amd64) echo "amd64" ;;
        arm64|aarch64) echo "arm64" ;;
        armv7l) echo "arm" ;;
        i386|i686) echo "386" ;;
        *) uname -m ;;
    esac
}

host_platform() {
    echo "$(host_os)_$(host_arch)"
}

ensure_dirs() {
    mkdir -p \
        "$OUTPUT_ABS/bin/platforms/$(host_os)/$(host_arch)" \
        "$OUTPUT_ABS/bin/tools/$(host_os)/$(host_arch)" \
        "$OUTPUT_ABS/logs" \
        "$OUTPUT_ABS/tmp/pids"
}

# Parse start-config.yml without external dependencies (supports both services:/tools: and serviceBinaries:/toolBinaries:).
parse_services() {
    awk '
        /^[[:space:]]*#/ || /^[[:space:]]*$/ { next }
        /^(services|serviceBinaries):[[:space:]]*$/ { section="services"; next }
        /^[^[:space:]]/ { section=""; next }
        section == "services" {
            line = $0
            sub(/#.*$/, "", line)
            gsub(/^[[:space:]]+|[[:space:]]+$/, "", line)
            if (line == "") next
            split(line, parts, ":")
            name = parts[1]
            count = parts[2]
            gsub(/^[[:space:]]+|[[:space:]]+$/, "", name)
            gsub(/^[[:space:]]+|[[:space:]]+$/, "", count)
            if (name != "" && count ~ /^[0-9]+$/ && count > 0) {
                print name, count
            }
        }
    ' "$START_CONFIG_ABS"
}

parse_tools() {
    awk '
        /^[[:space:]]*#/ || /^[[:space:]]*$/ { next }
        /^(tools|toolBinaries):[[:space:]]*$/ { section="tools"; next }
        /^[^[:space:]]/ { section=""; next }
        section == "tools" {
            line = $0
            sub(/#.*$/, "", line)
            gsub(/^[[:space:]]+-[[:space:]]+|[[:space:]]+$/, "", line)
            if (line != "") print line
        }
    ' "$START_CONFIG_ABS"
}

parse_max_fds() {
    awk '
        /^[[:space:]]*maxFileDescriptors:[[:space:]]*[0-9]+/ {
            sub(/#.*$/, "")
            split($0, parts, ":")
            gsub(/^[[:space:]]+|[[:space:]]+$/, "", parts[2])
            print parts[2]
            exit
        }
    ' "$START_CONFIG_ABS"
}

apply_fd_limit() {
    local max_fds
    max_fds="$(parse_max_fds || true)"
    if [[ -n "$max_fds" ]]; then
        ulimit -n "$max_fds" 2>/dev/null || true
    fi
}

discover_all_services() {
    local f dir
    for f in "$ROOT_DIR"/cmd/server-*/main.go "$ROOT_DIR"/cmd/server-rpc/server-*/main.go; do
        [[ -f "$f" ]] || continue
        dir="${f%/main.go}"
        printf '%s\n' "${dir#$ROOT_DIR/}"
    done
}

discover_all_tools() {
    local f dir
    for f in "$ROOT_DIR"/tools/*/main.go; do
        [[ -f "$f" ]] || continue
        dir="${f%/main.go}"
        printf '%s\n' "${dir#$ROOT_DIR/}"
    done
}

resolve_single_target() {
    local name="$1"
    if [[ -z "$name" || "$name" == "." || "$name" == ".." || "$name" == */* || "$name" == *\\* ]]; then
        log_error "invalid target \"$name\""
        return 1
    fi
    if [[ -f "$ROOT_DIR/tools/$name/main.go" ]]; then
        echo "tool:tools/$name"
    elif [[ -f "$ROOT_DIR/cmd/$name/main.go" ]]; then
        echo "service:cmd/$name"
    elif [[ -f "$ROOT_DIR/cmd/server-rpc/$name/main.go" ]]; then
        echo "service:cmd/server-rpc/$name"
    else
        log_error "unknown build target \"$name\""
        return 1
    fi
}

build_one() {
    local kind="$1"
    local rel_dir="$2"
    local goos="$3"
    local goarch="$4"

    local bin_name
    bin_name="$(basename "$rel_dir")"
    local ext=""
    if [[ "$goos" == "windows" ]]; then
        ext=".exe"
    fi

    local out_dir
    if [[ "$kind" == "tool" ]]; then
        out_dir="$OUTPUT_ABS/bin/tools/$goos/$goarch"
    else
        out_dir="$OUTPUT_ABS/bin/platforms/$goos/$goarch"
    fi
    mkdir -p "$out_dir"
    local out_path="$out_dir/${bin_name}${ext}"

    local -a ldflags=("-X" "server/v3/version.Version=${VERSION}")
    local -a args=("build" "-buildvcs=false" "-o" "$out_path")
    if [[ "$RELEASE" == "true" || "$RELEASE" == "1" ]]; then
        args+=("-trimpath" "-ldflags" "-s -w ${ldflags[*]}")
    else
        args+=("-ldflags" "${ldflags[*]}")
    fi
    args+=("./$rel_dir")

    log_info "Compiling $rel_dir -> ${out_path#$ROOT_DIR/} ($goos/$goarch)"
    (cd "$ROOT_DIR" && GOOS="$goos" GOARCH="$goarch" go "${args[@]}")
}

cmd_build() {
    local targets_arg="${1:-}"
    local -a service_dirs=()
    local -a tool_dirs=()

    if [[ -n "$targets_arg" ]]; then
        local name resolved
        for name in $targets_arg; do
            resolved="$(resolve_single_target "$name")" || return 1
            case "$resolved" in
                service:*) service_dirs+=("${resolved#service:}") ;;
                tool:*)    tool_dirs+=("${resolved#tool:}") ;;
            esac
        done
    else
        while IFS= read -r d; do
            [[ -n "$d" ]] && service_dirs+=("$d")
        done < <(discover_all_services)
        while IFS= read -r d; do
            [[ -n "$d" ]] && tool_dirs+=("$d")
        done < <(discover_all_tools)
    fi

    if [[ ${#service_dirs[@]} -eq 0 && ${#tool_dirs[@]} -eq 0 ]]; then
        log_error "at least one build target is required"
        return 1
    fi

    local platforms_list="$PLATFORMS"
    if [[ -z "$platforms_list" ]]; then
        platforms_list="$(host_platform)"
    fi
    platforms_list="${platforms_list//,/ }"

    local platform goos goarch
    for platform in $platforms_list; do
        goos="${platform%_*}"
        goarch="${platform#*_}"
        if [[ -z "$goos" || -z "$goarch" || "$goos" == "$platform" ]]; then
            log_error "invalid platform \"$platform\" (expected os_arch, e.g. linux_amd64)"
            return 1
        fi

        local active=0
        local -a pids=()
        local item kind rel_dir
        local -a queue=()
        for rel_dir in "${service_dirs[@]}"; do queue+=("service:$rel_dir"); done
        for rel_dir in "${tool_dirs[@]}";    do queue+=("tool:$rel_dir"); done

        for item in "${queue[@]}"; do
            kind="${item%%:*}"
            rel_dir="${item#*:}"
            build_one "$kind" "$rel_dir" "$goos" "$goarch" &
            pids+=($!)
            active=$((active + 1))
            if (( active >= JOBS )); then
                wait "${pids[0]}" || return 1
                pids=("${pids[@]:1}")
                active=$((active - 1))
            fi
        done
        local pid
        for pid in "${pids[@]}"; do
            wait "$pid" || return 1
        done
    done

    log_ok "Build completed successfully."
}

find_service_pids() {
    local bin_path="$1"
    ps -eo pid=,args= 2>/dev/null | awk -v target="$bin_path" '
        $2 == target { print $1 }
    '
}

cmd_stop() {
    if [[ ! -f "$START_CONFIG_ABS" ]]; then
        log_error "start config not found: $START_CONFIG_ABS"
        return 1
    fi
    local bin_dir="$OUTPUT_ABS/bin/platforms/$(host_os)/$(host_arch)"
    local -a services=()
    local name count
    while read -r name count; do
        [[ -n "$name" ]] && services+=("$name")
    done < <(parse_services)

    local pids_found=0
    for name in "${services[@]}"; do
        local bin_path="$bin_dir/$name"
        local pid
        while IFS= read -r pid; do
            [[ -n "$pid" ]] || continue
            pids_found=1
            kill -TERM "$pid" 2>/dev/null || true
        done < <(find_service_pids "$bin_path")
    done

    if (( pids_found == 0 )); then
        rm -f "$OUTPUT_ABS"/tmp/pids/*.pid 2>/dev/null || true
        log_ok "All services are already stopped."
        return 0
    fi

    local attempt
    for attempt in $(seq 1 15); do
        local still_running=()
        for name in "${services[@]}"; do
            local bin_path="$bin_dir/$name"
            if [[ -n "$(find_service_pids "$bin_path")" ]]; then
                still_running+=("$name")
            fi
        done
        if [[ ${#still_running[@]} -eq 0 ]]; then
            rm -f "$OUTPUT_ABS"/tmp/pids/*.pid 2>/dev/null || true
            log_ok "All services have been stopped."
            return 0
        fi
        log_warn "Waiting for services to stop (${attempt}/15): ${still_running[*]}"
        sleep 1
    done

    log_warn "Force killing remaining services..."
    for name in "${services[@]}"; do
        local bin_path="$bin_dir/$name"
        local pid
        while IFS= read -r pid; do
            [[ -n "$pid" ]] || continue
            kill -KILL "$pid" 2>/dev/null || true
        done < <(find_service_pids "$bin_path")
    done
    rm -f "$OUTPUT_ABS"/tmp/pids/*.pid 2>/dev/null || true
    log_ok "All services have been stopped."
}

cmd_check() {
    if [[ ! -f "$START_CONFIG_ABS" ]]; then
        log_error "start config not found: $START_CONFIG_ABS"
        return 1
    fi
    local bin_dir="$OUTPUT_ABS/bin/platforms/$(host_os)/$(host_arch)"
    local failed=0
    local name count
    while read -r name count; do
        [[ -n "$name" ]] || continue
        local bin_path="$bin_dir/$name"
        local -a pids=()
        local pid
        while IFS= read -r pid; do
            [[ -n "$pid" ]] && pids+=("$pid")
        done < <(find_service_pids "$bin_path")

        local actual=${#pids[@]}
        if (( actual != count )); then
            log_error "$name: expected $count running instance(s), found $actual"
            failed=1
        else
            log_ok "$name: $actual/$count running (PIDs: ${pids[*]})"
        fi
    done < <(parse_services)

    if (( failed != 0 )); then
        return 1
    fi

    log_info "Listening ports:"
    if command -v lsof >/dev/null 2>&1; then
        lsof -nP -iTCP -sTCP:LISTEN 2>/dev/null | grep -E "server-" || true
    elif command -v ss >/dev/null 2>&1; then
        ss -ltnp 2>/dev/null | grep -E "server-" || true
    fi
}

cmd_start() {
    if [[ ! -f "$START_CONFIG_ABS" ]]; then
        log_error "start config not found: $START_CONFIG_ABS"
        return 1
    fi
    ensure_dirs
    apply_fd_limit

    local bin_dir="$OUTPUT_ABS/bin/platforms/$(host_os)/$(host_arch)"
    local tool_dir="$OUTPUT_ABS/bin/tools/$(host_os)/$(host_arch)"

    log_info "Running pre-start tools..."
    local tool
    while IFS= read -r tool; do
        [[ -n "$tool" ]] || continue
        local tool_path="$tool_dir/$tool"
        if [[ ! -x "$tool_path" ]]; then
            log_error "Tool binary not found or not executable: $tool_path (run 'make build' first)"
            return 1
        fi
        log_info "Executing $tool -c $CONFIG_ABS"
        (cd "$tool_dir" && "$tool_path" -c "$CONFIG_ABS")
    done < <(parse_tools)
    log_ok "All pre-start tools executed successfully."

    cmd_stop

    log_info "Starting services..."
    local name count i
    while read -r name count; do
        [[ -n "$name" ]] || continue
        local bin_path="$bin_dir/$name"
        if [[ ! -x "$bin_path" ]]; then
            log_error "Service binary not found or not executable: $bin_path (run 'make build' first)"
            return 1
        fi
        for (( i = 0; i < count; i++ )); do
            local log_file="$OUTPUT_ABS/logs/${name}-${i}.log"
            local pid_file="$OUTPUT_ABS/tmp/pids/${name}-${i}.pid"
            log_info "Starting $name (index=$i)..."
            (cd "$bin_dir" && nohup "$bin_path" -i "$i" -c "$CONFIG_ABS" >> "$log_file" 2>&1 &
             echo $! > "$pid_file")
        done
    done < <(parse_services)

    sleep 1
    cmd_check
}

cmd_pack() {
    local out="$OUTPUT_ABS/server-linux-amd64.zip"
    if ! command -v python3 >/dev/null 2>&1; then
        log_error "python3 is required to package $out"
        return 1
    fi
    python3 - "$ROOT_DIR" "$out" <<'PY'
import os
import stat
import sys
import zipfile
from pathlib import Path

root = Path(sys.argv[1]).resolve()
out = Path(sys.argv[2]).resolve()

services = (
    "server-api",
    "server-rpc-user",
    "server-rpc-auth",
    "server-rpc-friend",
    "server-rpc-group",
    "server-rpc-msg",
    "server-rpc-conversation",
    "server-rpc-third",
    "server-msggateway",
    "server-msgtransfer",
    "server-push",
    "server-crontask",
    "server-cmdutils",
)
tools = ("check-component", "check-free-memory", "seq")

entries = (
    [f"_output/bin/platforms/linux/amd64/{s}" for s in services]
    + [f"_output/bin/tools/linux/amd64/{t}" for t in tools]
    + [
        "config",
        "docker-compose.yml",
        ".env",
        "start-config.yml",
        "version",
        "install.sh",
        "Makefile",
        "scripts/manage.sh",
        "LICENSE",
    ]
)

for entry in entries:
    full = root / entry
    if not full.exists():
        if "linux/amd64" in entry:
            sys.exit(f"{entry} missing; run: CGO_ENABLED=0 PLATFORMS=linux_amd64 make")
        sys.exit(f"{entry}: not found")

out.parent.mkdir(parents=True, exist_ok=True)
tmp = out.with_suffix(out.suffix + ".tmp")

def add_file(zf: zipfile.ZipFile, src: Path, arcname: str) -> None:
    st = src.stat()
    zi = zipfile.ZipInfo.from_file(src, arcname)
    zi.compress_type = zipfile.ZIP_DEFLATED
    mode = stat.S_IFREG | (0o755 if (arcname.startswith("_output/bin/") or arcname in ("scripts/manage.sh", "install.sh")) else (stat.S_IMODE(st.st_mode) or 0o644))
    zi.external_attr = mode << 16
    with src.open("rb") as f:
        zf.writestr(zi, f.read())

try:
    with zipfile.ZipFile(tmp, "w", compression=zipfile.ZIP_DEFLATED) as zf:
        for entry in entries:
            full = root / entry
            if full.is_file():
                add_file(zf, full, entry)
            else:
                for path in sorted(full.rglob("*")):
                    if path.name == ".DS_Store" or path.is_dir():
                        continue
                    arcname = path.relative_to(root).as_posix()
                    add_file(zf, path, arcname)
    os.replace(tmp, out)
except Exception:
    if tmp.exists():
        tmp.unlink()
    raise

print(f"wrote {out}")
PY
}

usage() {
    cat <<EOF
Usage: $0 <command> [args...]

Commands:
  build [targets]   Compile all services & tools, or only space-separated targets
  start             Run pre-start tools and start all configured services
  stop              Stop all running services
  restart           Stop then start all services
  check             Verify running service counts and listening ports
  pack              Build portable _output/server-linux-amd64.zip release archive
EOF
}

main() {
    local cmd="${1:-}"
    shift || true
    case "$cmd" in
        build)   cmd_build "${*:-}" ;;
        start)   cmd_start ;;
        stop)    cmd_stop ;;
        restart) cmd_stop && cmd_start ;;
        check)   cmd_check ;;
        pack)    cmd_pack ;;
        -h|--help|"") usage ;;
        *) log_error "unknown command: $cmd"; usage >&2; exit 1 ;;
    esac
}

main "$@"
