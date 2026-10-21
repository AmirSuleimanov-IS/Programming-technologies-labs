#!/usr/bin/env bash
# Запускает инструмент из frontend/node_modules/.bin внутри папки frontend.
# pre-commit передаёт пути от корня репозитория (frontend/src/app.tsx) — префикс убирается.
set -euo pipefail
cd "$(dirname "$0")/../frontend"
tool="$1"
shift
args=()
for arg in "$@"; do
  args+=("${arg#frontend/}")
done
exec "node_modules/.bin/$tool" "${args[@]}"
