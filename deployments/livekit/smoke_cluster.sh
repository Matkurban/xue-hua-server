#!/usr/bin/env bash
# Conference / LiveKit cluster smoke helpers (N=1 and N=2).
# Does not start containers by default — validates config then prints checklist.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")" && pwd)"
cd "$ROOT"

fail=0
check() {
  local desc="$1"; shift
  if "$@"; then
    echo "OK  $desc"
  else
    echo "FAIL $desc"
    fail=1
  fi
}

echo "== Config checks =="
check "livekit.yaml has redis" grep -q '^redis:' livekit.yaml
check "livekit.yaml has conference webhook" grep -q 'conference/livekit_webhook' livekit.yaml
check "livekit.yaml has meeting webhook" grep -q 'meeting/livekit_webhook' livekit.yaml
check "N=1 compose present" test -f docker-compose.cluster.yml
check "N=2 compose present" test -f docker-compose.cluster.n2.yml
check "N=2 compose has livekit-a" grep -q 'livekit-a' docker-compose.cluster.n2.yml
check "N=2 compose has livekit-b" grep -q 'livekit-b' docker-compose.cluster.n2.yml
check "ingress whip_port" grep -q 'whip_port' ingress.yaml
check "k8s webhook paths" grep -q 'conference/livekit_webhook' k8s/livekit.yaml

echo
echo "== Manual checklist (run after compose up) =="
cat <<'EOF'
N=1:
  1. docker compose -f docker-compose.cluster.yml up -d
  2. Create conference → join as host + 2 listeners (different devices/simulators)
  3. Host: mute one user, raise/approve hand, start/stop egress, create RTMP+WHIP ingress
  4. Leave/end → Redis keys conference:* gone; LiveKit room deleted

N=2:
  1. docker compose -f docker-compose.cluster.n2.yml up -d
  2. Host connects via ws://host:7880; invitee via ws://host:7980 (same roomID)
  3. Role change / kick / conference_sync still applied on both clients
  4. Kill livekit-a; new joins via :7980 still see room if redis holds state
     (existing media sessions on dead node will drop — expected)
  5. chat-api multi-replica: create on pod A, join/get on pod B (shared Redis)

Cleanup:
  docker compose -f docker-compose.cluster.yml down -v
  docker compose -f docker-compose.cluster.n2.yml down -v
EOF

if [[ "${1:-}" == "--up-n1" ]]; then
  docker compose -f docker-compose.cluster.yml up -d
elif [[ "${1:-}" == "--up-n2" ]]; then
  docker compose -f docker-compose.cluster.n2.yml up -d
fi

exit "$fail"
