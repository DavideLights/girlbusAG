#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

mkdir -p dist
echo "[girlbusAG] Building Linux native binary..."
go build -ldflags="-s -w" -o dist/girlbus_linux .
chmod +x dist/girlbus_linux
echo "[girlbusAG] Done: dist/girlbus_linux"
