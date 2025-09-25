#!/usr/bin/env bash
set -euo pipefail

# Basit DNS ve IPv6 sızıntı testi.
# Gereksinimler: dig, curl, jq

TARGET_DOMAIN="resolver.dnscrypt.info"
DNS_SERVER="9.9.9.9"

printf "==> DNS test\n"
dig @${DNS_SERVER} ${TARGET_DOMAIN} > /tmp/leaktest.dig

printf "==> IPv6 test\n"
if curl -6 --silent https://ifconfig.co > /tmp/leaktest.ipv6; then
  echo "IPv6 cevap:" $(cat /tmp/leaktest.ipv6)
else
  echo "IPv6 erişimi yok"
fi

echo "Loglar /tmp/leaktest.* altında"
