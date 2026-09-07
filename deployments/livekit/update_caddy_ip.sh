#!/usr/bin/env bash
# Fill the host's primary IPv4 into config/caddy.yaml TURN upstream (replaces HOST_IP or a previous IP).
set -euo pipefail
cfg="${1:-$(cd "$(dirname "$0")/../.." && pwd)/config/caddy.yaml}"
ip=$(ip addr show | grep "inet " | grep -v 127.0.0. | head -1 | cut -d" " -f6 | cut -d/ -f1)
sed -i.orig -E "s/\"(HOST_IP|[0-9.]+)(:5349)\"/\"$ip\2\"/" "$cfg"
echo "caddy TURN upstream -> $ip:5349 ($cfg)"
