#!/bin/sh
set -eu
HERE=$(cd "$(dirname "$0")" && pwd)

echo "Captive Wi-Fi: installing background service..."
install -m 0644 "$HERE/systemd/rcp-daemon.service" /etc/systemd/system/rcp-daemon.service
install -m 0644 "$HERE/systemd/rcp-portal.service" /etc/systemd/system/rcp-portal.service
systemctl daemon-reload
systemctl enable --now rcp-daemon.service

echo "Captive Wi-Fi enabled."
echo "Status: systemctl status rcp-daemon.service"
exit 0
