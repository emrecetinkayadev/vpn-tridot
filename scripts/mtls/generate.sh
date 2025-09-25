#!/usr/bin/env bash
set -euo pipefail

# Generates CA, server, and client certificates for a given environment (staging/prod/local).
# Usage: ./generate.sh staging api.vpn.example.com agent.vpn.example.com
# Certificates are written under scripts/mtls/build/<env>/

ENVIRONMENT="${1:-}"
if [[ -z "${ENVIRONMENT}" ]]; then
  echo "Usage: $0 <environment> [server-hostname ...]" >&2
  exit 1
fi

shift
HOSTNAMES=("${@}")
if [[ ${#HOSTNAMES[@]} -eq 0 ]]; then
  HOSTNAMES=("api.${ENVIRONMENT}.vpn.internal")
fi

BUILD_DIR="$(dirname "$0")/build/${ENVIRONMENT}"
mkdir -p "${BUILD_DIR}"

CA_KEY="${BUILD_DIR}/ca.key.pem"
CA_CRT="${BUILD_DIR}/ca.crt.pem"
SERVER_KEY="${BUILD_DIR}/server.key.pem"
SERVER_CSR="${BUILD_DIR}/server.csr.pem"
SERVER_CRT="${BUILD_DIR}/server.crt.pem"
CLIENT_KEY="${BUILD_DIR}/client.key.pem"
CLIENT_CSR="${BUILD_DIR}/client.csr.pem"
CLIENT_CRT="${BUILD_DIR}/client.crt.pem"

if [[ ! -f "${CA_KEY}" ]]; then
  echo "[*] Generating new CA for ${ENVIRONMENT}" >&2
  openssl req -x509 -new -nodes \
    -newkey rsa:4096 \
    -days 730 \
    -keyout "${CA_KEY}" \
    -out "${CA_CRT}" \
    -subj "/CN=TriDot VPN ${ENVIRONMENT} mTLS CA"
else
  echo "[+] Reusing existing CA at ${CA_CRT}" >&2
fi

SAN_CONFIG="${BUILD_DIR}/server-san.cnf"
{
  echo "[ req ]"
  echo "distinguished_name = req"
  echo "req_extensions = v3_req"
  echo "[ v3_req ]"
  echo "keyUsage = critical, digitalSignature, keyEncipherment"
  echo "extendedKeyUsage = serverAuth"
  echo -n "subjectAltName = @alt_names\n[ alt_names ]\n"
  idx=1
  for host in "${HOSTNAMES[@]}"; do
    echo "DNS.${idx} = ${host}"
    ((idx++))
  done
} > "${SAN_CONFIG}"

if [[ ! -f "${SERVER_KEY}" ]]; then
  echo "[*] Generating server key/cert for hosts: ${HOSTNAMES[*]}" >&2
  openssl req -new -nodes -newkey rsa:4096 \
    -keyout "${SERVER_KEY}" \
    -out "${SERVER_CSR}" \
    -subj "/CN=${HOSTNAMES[0]}" \
    -config "${SAN_CONFIG}"
  openssl x509 -req -in "${SERVER_CSR}" -CA "${CA_CRT}" -CAkey "${CA_KEY}" \
    -CAcreateserial -out "${SERVER_CRT}" -days 365 \
    -extensions v3_req -extfile "${SAN_CONFIG}"
fi

CLIENT_CONFIG="${BUILD_DIR}/client-ext.cnf"
cat > "${CLIENT_CONFIG}" <<'INI'
keyUsage = critical, digitalSignature
extendedKeyUsage = clientAuth
INI

if [[ ! -f "${CLIENT_KEY}" ]]; then
  echo "[*] Generating client key/cert" >&2
  openssl req -new -nodes -newkey rsa:4096 \
    -keyout "${CLIENT_KEY}" \
    -out "${CLIENT_CSR}" \
    -subj "/CN=tridot-agent"
  openssl x509 -req -in "${CLIENT_CSR}" -CA "${CA_CRT}" -CAkey "${CA_KEY}" \
    -CAcreateserial -out "${CLIENT_CRT}" -days 365 \
    -extfile "${CLIENT_CONFIG}"
fi

echo "[✓] Certificates written to ${BUILD_DIR}" >&2
