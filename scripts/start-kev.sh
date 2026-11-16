#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
KEV_DIR="$ROOT_DIR/kev-upstream"
UV_BIN="${UV_BIN:-$HOME/.local/bin/uv}"

if [[ ! -d "$KEV_DIR" ]]; then
  echo "KEV no está instalado. Ejecuta primero ./scripts/install-kev.sh" >&2
  exit 1
fi

export KEV_API_KEY="${KEV_API_KEY:-local-demo-key}"
export KEV_PREFIX_CACHE="${KEV_PREFIX_CACHE:-8}"
export KEV_DATE_FACTS="${KEV_DATE_FACTS:-1}"
# Xet puede quedar bloqueado reconstruyendo archivos grandes en algunas redes.
# La descarga HTTP estándar es más lenta, pero más predecible para una demo local.
export HF_HUB_DISABLE_XET="${HF_HUB_DISABLE_XET:-1}"

exec "$UV_BIN" run --directory "$KEV_DIR" --extra serve \
  python -m kev.serve \
  --run "jaredpalmer/kev-0.8b@v1.0" \
  --host 0.0.0.0 \
  --port 8009
