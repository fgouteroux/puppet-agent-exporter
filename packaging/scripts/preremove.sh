#!/bin/sh
set -e

if [ -d /run/systemd/system ]; then
	systemctl --no-reload disable --now puppet-agent-exporter.service >/dev/null 2>&1 || true
fi
