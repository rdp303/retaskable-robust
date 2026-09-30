#!/bin/sh
set -eu
ROOT=$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)
mkdir -p "$ROOT/package/bin"
(
  cd "$ROOT/daemon"
  CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -ldflags='-s -w' -o "$ROOT/package/bin/rcp-daemon" .
)
chmod +x "$ROOT/package/enable.sh" "$ROOT/package/disable.sh" "$ROOT/package/portal-session.sh" "$ROOT/package/diagnostics.sh" "$ROOT/package/bin/rcp-daemon"
echo "Built AppLoad folder: $ROOT/package"
