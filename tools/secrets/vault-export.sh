#!/usr/bin/env bash
set -euo pipefail

if [[ -z "${VAULT_ADDR:-}" ]]; then
  echo "VAULT_ADDR is required" >&2
  exit 1
fi
if [[ -z "${VAULT_TOKEN:-}" ]]; then
  echo "VAULT_TOKEN is required" >&2
  exit 1
fi
if [[ -z "${VAULT_PATH:-}" ]]; then
  echo "VAULT_PATH is required" >&2
  exit 1
fi

TARGET_FILE=${TARGET_FILE:-/run/secrets/vault.env}

response=$(curl -sSf -H "X-Vault-Token: ${VAULT_TOKEN}" "${VAULT_ADDR%/}/v1/${VAULT_PATH#*/}")

mkdir -p "$(dirname "${TARGET_FILE}")"

if command -v jq >/dev/null 2>&1; then
  echo "$response" | jq -r '.data.data | to_entries[] | "\(.key)=\(.value)"' > "${TARGET_FILE}"
else
  echo "$response" > "${TARGET_FILE}"
fi
