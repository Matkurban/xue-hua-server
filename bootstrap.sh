#!/usr/bin/env bash
set -euo pipefail

if ! command -v go >/dev/null 2>&1; then
    echo "Go is not installed. Please install Go 1.26+ first." >&2
    exit 1
fi

if ! command -v make >/dev/null 2>&1; then
    echo "GNU Make is not installed. Please install make first (or run ./scripts/manage.sh directly)." >&2
    exit 1
fi

chmod +x scripts/manage.sh
echo "Environment verified (go + make)."
go mod download
