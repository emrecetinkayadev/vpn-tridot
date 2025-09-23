#!/usr/bin/env bash
set -euo pipefail

if [[ -z "${SOPS_SOURCE:-}" ]]; then
  echo "SOPS_SOURCE variable is required" >&2
  exit 1
fi

if [[ -z "${SOPS_TARGET:-}" ]]; then
  echo "SOPS_TARGET variable is required" >&2
  exit 1
fi

mkdir -p "$(dirname "${SOPS_TARGET}")"

sops -d "${SOPS_SOURCE}" > "${SOPS_TARGET}"
