#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")"/.. && pwd)"

pushd "$ROOT_DIR" > /dev/null

if command -v pre-commit >/dev/null 2>&1; then
  echo "Installing pre-commit hooks..."
  pre-commit install
else
  echo "pre-commit not found. Install it and rerun this script." >&2
fi

if command -v pnpm >/dev/null 2>&1; then
  echo "Installing frontend dependencies..."
  pnpm install --dir frontend
else
  echo "pnpm not found. Skipping frontend dependency install." >&2
fi

if command -v go >/dev/null 2>&1; then
  echo "Tidying Go modules (if present)..."
  find backend node-agent -maxdepth 1 -name go.mod -execdir go mod tidy \;
else
  echo "Go compiler not found. Skipping go mod tidy." >&2
fi

popd > /dev/null
