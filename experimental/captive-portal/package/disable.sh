#!/bin/sh
set -eu
systemctl disable --now rcp-daemon.service 2>/dev/null || true
systemctl stop rcp-portal.service 2>/dev/null || true
rm -f /etc/systemd/system/rcp-daemon.service /etc/systemd/system/rcp-portal.service
systemctl daemon-reload
systemctl start xochitl 2>/dev/null || true
echo "Captive Wi-Fi disabled."
