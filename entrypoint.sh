#!/bin/bash
set -euxo pipefail

# Cleanup to be "stateless" on startup, otherwise pulseaudio daemon can't start
rm -rf /var/run/pulse /var/lib/pulse /root/.config/pulse

# Start pulseaudio as system wide daemon; for debugging it helps to start in non-daemon mode
pulseaudio -D --verbose --exit-idle-time=-1 --system --disallow-exit

pactl load-module module-null-sink sink_name="grab" sink_properties=device.description="monitorOUT"

# Start xvfb
export DISPLAY=:10
Xvfb :10 -ac -screen 0 1024x768x16 & node home/node/src/record