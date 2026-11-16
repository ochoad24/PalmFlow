#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
KEV_DIR="$ROOT_DIR/kev-upstream"
UV_BIN="${UV_BIN:-$HOME/.local/bin/uv}"

if [[ ! -x "$UV_BIN" ]]; then
  echo "Instalando uv en $HOME/.local/bin..."
  curl -LsSf https://astral.sh/uv/install.sh | sh
fi

if [[ ! -d "$KEV_DIR/.git" ]]; then
  git clone https://github.com/jaredpalmer/kev.git "$KEV_DIR"
fi

git -C "$KEV_DIR" fetch --tags origin
git -C "$KEV_DIR" checkout kev-1.0
"$UV_BIN" sync --directory "$KEV_DIR" --extra serve

cat <<EOF

KEV quedó instalado en:
  $KEV_DIR

Inícialo con:
  ./scripts/start-kev.sh

La primera ejecución descargará Kev-0.8B y Qwen3.5-0.8B.
EOF
