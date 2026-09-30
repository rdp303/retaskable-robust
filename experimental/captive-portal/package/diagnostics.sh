#!/bin/sh
set -u

RUNTIME=${RCP_RUNTIME_DIR:-/run/remarkable-captive-portal}

echo '=== reMarkable Captive Wi-Fi diagnostics ==='
date -u 2>/dev/null || true
uname -a 2>/dev/null || true

echo
echo '--- daemon status ---'
systemctl status rcp-daemon.service --no-pager 2>&1 || true

echo
echo '--- portal status ---'
systemctl status rcp-portal.service --no-pager 2>&1 || true

echo
echo '--- current detector state ---'
cat "$RUNTIME/state.json" 2>/dev/null || echo '(no state.json)'

echo
echo '--- discovered portal URL ---'
cat "$RUNTIME/portal-url" 2>/dev/null || echo '(no portal-url)'

echo
echo '--- recent daemon logs ---'
journalctl -u rcp-daemon.service -n 120 --no-pager 2>&1 || true

echo
echo '--- recent portal logs ---'
journalctl -u rcp-portal.service -n 120 --no-pager 2>&1 || true
