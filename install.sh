#!/usr/bin/env bash

# Copyright © 2023 OpenIM. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -euo pipefail

usage() {
    echo "Usage: $0 LOCAL_RELEASE.zip_OR_DIRECTORY NEW_INSTALL_DIRECTORY"
    echo "Installs a locally built server Linux/amd64 release; never downloads or starts services."
}

fail() {
    echo "server install: $*" >&2
    exit 1
}

if [[ "${1:-}" == "--help" || "${1:-}" == "-h" ]]; then
    usage
    exit 0
fi
[[ $# -eq 2 ]] || { usage >&2; exit 1; }
source_path="$1"
destination="$2"
[[ -e "$source_path" ]] || fail "release does not exist: $source_path"
[[ ! -e "$destination" && ! -L "$destination" ]] || fail "destination already exists; refusing to overwrite"
command -v python3 >/dev/null || fail "python3 is required to validate and install the local release"

python3 - "$source_path" "$destination" <<'PY'
import shutil
import stat
import sys
import zipfile
from pathlib import Path, PurePosixPath

source = Path(sys.argv[1]).resolve()
destination = Path(sys.argv[2]).absolute()
services = (
    "api", "rpc-user", "rpc-auth", "rpc-friend", "rpc-group", "rpc-msg",
    "rpc-conversation", "rpc-third", "msggateway", "msgtransfer", "push", "crontask", "cmdutils",
)
required = [
    "start-config.yml", "Makefile", "scripts/manage.sh",
    "version/version.go", "LICENSE",
] + [f"_output/bin/platforms/linux/amd64/server-{name}" for name in services]
required += [f"_output/bin/tools/linux/amd64/{name}" for name in ("check-component", "check-free-memory", "seq")]
required += [f"config/server-{name}.yml" for name in services if name != "cmdutils"]
required += [f"config/{name}.yml" for name in (
    "discovery", "kafka", "local-cache", "log", "rustfs", "mongodb", "notification", "redis", "share", "webhooks",
)]

def validate_path(name):
    path = PurePosixPath(name)
    if not name or path.is_absolute() or ".." in path.parts or "\\" in name or str(path) != name.rstrip("/"):
        raise ValueError(f"unsafe release path: {name}")

created = False
try:
    if source.is_dir():
        if destination == source or source in destination.parents:
            raise ValueError("destination cannot be inside the source release")
        missing = [name for name in required if not (source / name).is_file()]
        if missing:
            raise ValueError("missing required release files: " + ", ".join(missing))
        for path in source.rglob("*"):
            if path.is_symlink() or not (path.is_file() or path.is_dir()):
                raise ValueError(f"unsupported release entry: {path.relative_to(source)}")
        destination.mkdir(mode=0o700, parents=True, exist_ok=False)
        created = True
        shutil.copytree(source, destination, dirs_exist_ok=True)
    else:
        with zipfile.ZipFile(source) as archive:
            entries = archive.infolist()
            names = set()
            for entry in entries:
                validate_path(entry.filename)
                if entry.filename in names:
                    raise ValueError(f"duplicate release entry: {entry.filename}")
                names.add(entry.filename)
                mode = entry.external_attr >> 16
                if stat.S_ISLNK(mode) or stat.S_IFMT(mode) not in (0, stat.S_IFREG, stat.S_IFDIR):
                    raise ValueError(f"unsupported release entry: {entry.filename}")
            missing = [name for name in required if name not in names or archive.getinfo(name).is_dir()]
            if missing:
                raise ValueError("missing required release files: " + ", ".join(missing))
            destination.mkdir(mode=0o700, parents=True, exist_ok=False)
            created = True
            archive.extractall(destination)
    destination.chmod(0o700)
    launcher = destination / "scripts/manage.sh"
    launcher.chmod(launcher.stat().st_mode | stat.S_IXUSR | stat.S_IXGRP | stat.S_IXOTH)
    for path in (destination / "_output/bin").rglob("*"):
        if path.is_file():
            path.chmod(path.stat().st_mode | stat.S_IXUSR | stat.S_IXGRP | stat.S_IXOTH)
except (OSError, ValueError, UnicodeError, zipfile.BadZipFile, RuntimeError) as error:
    if created:
        shutil.rmtree(destination)
    sys.exit(f"server install: {error}")
print(f"Installed local release to {destination}")
PY

echo "Review config/ and start the required infrastructure before starting."
echo "On Linux/amd64: cd \"$destination\" && make start (or ./scripts/manage.sh start)"
