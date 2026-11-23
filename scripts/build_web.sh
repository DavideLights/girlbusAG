#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

mkdir -p dist/web

echo "[girlbusAG] Building WebAssembly binary..."
GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o dist/web/girlbus.wasm .

# Locate wasm_exec.js from Go's GOROOT
GOROOT_DIR="$(go env GOROOT)"
WASM_EXEC=""
for candidate in "$GOROOT_DIR/misc/wasm/wasm_exec.js" "$GOROOT_DIR/lib/wasm/wasm_exec.js"; do
  if [[ -f "$candidate" ]]; then
    WASM_EXEC="$candidate"
    break
  fi
done

if [[ -n "$WASM_EXEC" ]]; then
  cp "$WASM_EXEC" dist/web/wasm_exec.js
else
  echo "[girlbusAG] Warning: wasm_exec.js not found in GOROOT ($GOROOT_DIR)"
fi

if command -v brotli >/dev/null 2>&1; then
  echo "[girlbusAG] Compressing girlbus.wasm with brotli..."
  brotli -f -q 11 dist/web/girlbus.wasm -o dist/web/girlbus.wasm.br
fi

cat > dist/web/index.html <<'EOF'
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0, user-scalable=no" />
  <title>GIRLBUS: NEON STOP (girlbusAG v2.0)</title>
  <style>
    html, body {
      margin: 0;
      padding: 0;
      width: 100%;
      height: 100%;
      background: #0d0814;
      overflow: hidden;
      display: flex;
      align-items: center;
      justify-content: center;
    }
    canvas {
      image-rendering: pixelated;
      image-rendering: crisp-edges;
    }
  </style>
  <script src="wasm_exec.js"></script>
  <script>
    const go = new Go();
    WebAssembly.instantiateStreaming(fetch("girlbus.wasm"), go.importObject).then((result) => {
      go.run(result.instance);
    });
  </script>
</head>
<body></body>
</html>
EOF

echo "[girlbusAG] Done: dist/web/"
