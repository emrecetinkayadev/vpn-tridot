#!/usr/bin/env bash
set -euo pipefail

if [[ -z "${GOOSE_DBSTRING:-}" ]]; then
  echo "GOOSE_DBSTRING is required" >&2
  exit 1
fi

if [[ -z "${GOOSE_DRIVER:-postgres}" ]]; then
  export GOOSE_DRIVER=postgres
fi

pushd "$(dirname "${BASH_SOURCE[0]}")/../../backend" > /dev/null

goose up

popd > /dev/null
