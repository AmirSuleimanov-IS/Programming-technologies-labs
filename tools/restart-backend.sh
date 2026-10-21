#!/usr/bin/env bash
# Перезапускает локальный бэкенд (go run, режим development) с логом в tools/raw/backend_dev.log
set -euo pipefail
cd "$(dirname "$0")/.."
pid=$(netstat -ano | grep ":8090 .*LISTENING" | awk '{print $5}' | head -1 || true)
if [ -n "$pid" ]; then taskkill //F //PID "$pid" > /dev/null; fi
cd backend
go build -o ../tools/raw/api.exe ./cmd/api
(../tools/raw/api.exe > ../tools/raw/backend_dev.log 2>&1 &)
for _ in $(seq 1 30); do
  curl -sf http://127.0.0.1:8090/api/health > /dev/null && { echo "backend ready"; exit 0; }
  sleep 1
done
echo "backend did not start"; tail -5 ../tools/raw/backend_dev.log; exit 1
