#!/usr/bin/env bash
set -euo pipefail

TIMEOUT=120

echo "Waiting $TIMEOUT seconds for SSH connections..."
sleep "$TIMEOUT"

# Check for active SSH sessions
if who | grep -q "pts/"; then
  echo "SSH client detected, keeping WiFi enabled"
  exit 0
fi

echo "No SSH clients detected, disabling WiFi"

# NetworkManager
if command -v nmcli >/dev/null; then
  nmcli radio wifi off
  exit 0
fi

# systemd-networkd / rfkill fallback
if command -v rfkill >/dev/null; then
  rfkill block wifi
  exit 0
fi

echo "Could not disable WiFi: no known tool found"
exit 1
