#!/usr/bin/env bash
# golangci-lint проверяет весь модуль backend (линтеры unused и unparam видят только пакет целиком)
set -euo pipefail
cd "$(dirname "$0")/../backend"
exec golangci-lint run ./...
