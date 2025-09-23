# Secrets Scripts

Bu klasör, SOPS şifreli dosyaları ile HashiCorp Vault arasında hızlı geçiş yapmayı kolaylaştırır.

## sops-decrypt.sh

```
#!/usr/bin/env bash
set -euo pipefail

: "${SOPS_SOURCE:?SOPS_SOURCE is required}"
: "${SOPS_TARGET:?SOPS_TARGET is required}"

mkdir -p "$(dirname "${SOPS_TARGET}")"
sops -d "${SOPS_SOURCE}" > "${SOPS_TARGET}"
```

Kullanım:

```bash
SOPS_SOURCE=secrets/backend.enc.json \
SOPS_TARGET=/run/secrets/backend.json \
./tools/secrets/sops-decrypt.sh
```

Script, SOPS CLI'nin ortamda kurulu olduğunu varsayar.

## vault-export.sh

```
#!/usr/bin/env bash
set -euo pipefail

: "${VAULT_ADDR:?}"
: "${VAULT_TOKEN:?}"
: "${VAULT_PATH:?}"
TARGET_FILE=${TARGET_FILE:-/run/secrets/vault.env}

response=$(curl -sSf -H "X-Vault-Token: ${VAULT_TOKEN}" "${VAULT_ADDR%/}/v1/${VAULT_PATH#*/}")
if command -v jq >/dev/null 2>&1; then
  echo "$response" | jq -r '.data.data | to_entries[] | "\(.key)=\(.value)"' > "${TARGET_FILE}"
else
  echo "$response" > "${TARGET_FILE}"
fi
```

Kullanım:

```bash
VAULT_ADDR=https://vault.internal.example.com \
VAULT_TOKEN=$(cat ~/.vault-token) \
VAULT_PATH=kv/data/vpn-backend/prod \
TARGET_FILE=/run/secrets/vault.env \
./tools/secrets/vault-export.sh
```

Bu dosya `.env` formatında çıktıyı yazar; çalışma zamanında `SOPS_SECRETS_ENABLED=true` + `SOPS_SECRETS_FORMAT=env` ile kullanılabilir.
