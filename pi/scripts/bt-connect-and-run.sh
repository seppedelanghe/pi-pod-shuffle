#!/usr/bin/env bash
set -euo pipefail

MAC="AA:BB:CC:DD:EE:FF" # update your headphones MAC
CMD="/home/pipod/pi-pod-shuffle /home/pipod/FLACs/library.json"

# Wait for bluetoothd
echo "Waiting for bluetoothd..."
until systemctl is-active --quiet bluetooth; do
  sleep 1
done

# Wait for controller
until bluetoothctl list | grep -q Controller; do
  sleep 1
done

echo "Bluetooth is up"

# Try connecting until success
while true; do
  echo "Trying to connect to $MAC..."
  if bluetoothctl connect "$MAC" | grep -q "Connection successful"; then
    echo "Connected to $MAC"

    echo "Waiting for ALSA btheadset PCM..."
    until aplay -L | grep -q "^btheadset$"; do
      sleep 1
    done

    echo "ALSA btheadset ready, starting app"
    exec $CMD
  fi

  sleep 3
done
