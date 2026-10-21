#!/usr/bin/env bash
# Перезапускает локальный бэкенд (режим development) с логом в tools/raw/backend_dev.log.
# Останавливается только собственный tools/raw/api.exe. Если порт 8090 занят чем-то другим
# (например, прокси портов Docker Desktop для контейнера backend), скрипт не трогает этот
# процесс — его завершение роняет Docker Desktop — и сообщает об ошибке.
set -euo pipefail
cd "$(dirname "$0")/.."
taskkill //F //IM api.exe > /dev/null 2>&1 || true
sleep 1
if netstat -ano | grep -q ":8090 .*LISTENING"; then
  echo "port 8090 is busy (docker compose backend?): run 'docker compose stop backend' first"
  exit 1
fi
cd backend
go build -o ../tools/raw/api.exe ./cmd/api
(../tools/raw/api.exe > ../tools/raw/backend_dev.log 2>&1 &)
for _ in $(seq 1 30); do
  curl -sf http://127.0.0.1:8090/api/health > /dev/null && { echo "backend ready"; exit 0; }
  sleep 1
done
echo "backend did not start"; tail -5 ../tools/raw/backend_dev.log; exit 1
