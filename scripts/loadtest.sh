#!/usr/bin/env bash
set -euo pipefail

# WireGuard tüneli üzerinden iperf3 ile throughput testi.
# Kullanım: ./scripts/loadtest.sh <server-ip> [duration]

SERVER=${1:-}
DURATION=${2:-30}

if [[ -z "${SERVER}" ]]; then
  echo "Usage: $0 <server-ip> [duration]" >&2
  exit 1
fi

iperf3 -c "${SERVER}" -u -b 0 -t "${DURATION}"
