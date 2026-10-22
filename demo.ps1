# Запуск SportReg для показа лабораторных работ (например, после перезапуска компьютера).
# Из папки technologies:  powershell -ExecutionPolicy Bypass -File demo.ps1 [-Open] [-Traffic]
#   -Open     открыть приложение, Grafana и Prometheus в браузере
#   -Traffic  2 минуты подавать нагрузку, чтобы на дашборде Grafana появились графики
# Остановить всё: docker compose stop
param([switch]$Open, [switch]$Traffic)

$ErrorActionPreference = 'Continue' # ошибки внешних команд проверяются по $LASTEXITCODE
Set-Location $PSScriptRoot

function Test-DockerEngine {
    # пока Docker Desktop запускается, docker info может зависнуть — поэтому с таймаутом
    $job = Start-Job { docker info --format '{{.ServerVersion}}' 2>$null }
    $version = if (Wait-Job $job -Timeout 10) { Receive-Job $job } else { $null }
    Remove-Job $job -Force
    return [bool]("$version" -match '^\d+\.')
}

function Step($text) { Write-Host $text -ForegroundColor Cyan }
function Fail($text) { Write-Host "ОШИБКА: $text" -ForegroundColor Red; exit 1 }

Step '1/4 Docker Desktop'
if (-not (Test-DockerEngine)) {
    if (-not (Get-Process 'Docker Desktop' -ErrorAction SilentlyContinue)) {
        Start-Process 'C:\Program Files\Docker\Docker\Docker Desktop.exe'
    }
    Write-Host '    запускается, ждём (до 3 минут)...'
    $deadline = (Get-Date).AddMinutes(3)
    while (-not (Test-DockerEngine)) {
        if ((Get-Date) -gt $deadline) { Fail 'Docker Desktop не запустился за 3 минуты. Откройте его вручную и запустите скрипт снова.' }
        Start-Sleep -Seconds 5
    }
}
Write-Host '    Docker работает'

Step '2/4 Контейнеры: PostgreSQL, Redis, backend, frontend, Prometheus, Grafana'
docker compose up -d
if ($LASTEXITCODE -ne 0) { Fail 'docker compose up завершился с ошибкой' }

Step '3/4 Ждём готовности API'
$deadline = (Get-Date).AddMinutes(2)
while ($true) {
    try {
        $health = Invoke-RestMethod 'http://127.0.0.1:8090/api/health' -TimeoutSec 3
        if ($health.data.status -eq 'ok') { break }
    } catch { }
    if ((Get-Date) -gt $deadline) { Fail 'API не ответил за 2 минуты. Смотрите: docker compose logs backend' }
    Start-Sleep -Seconds 2
}
Write-Host '    API отвечает'

Step '4/4 Демо-пользователи (создаются, если их ещё нет)'
node tools/api-demo.mjs tools/steps/demo_users.json
if ($LASTEXITCODE -ne 0) { Fail 'не удалось создать демо-пользователей (нужен Node.js)' }
Write-Host '    готово'

if ($Traffic) {
    Start-Process node -ArgumentList 'traffic.mjs', '120' -WorkingDirectory "$PSScriptRoot\tools" -WindowStyle Minimized
    Write-Host '    нагрузка для Grafana запущена на 2 минуты'
}

Write-Host ''
Write-Host 'Всё запущено:' -ForegroundColor Green
Write-Host '  Приложение   http://localhost:8091'
Write-Host '  API          http://localhost:8090/api/health'
Write-Host '  Метрики      http://localhost:8090/metrics'
Write-Host '  Prometheus   http://localhost:9091/targets'
Write-Host '  Grafana      http://localhost:3001  (admin / admin, дашборд «SportReg — API»)'
Write-Host ''
Write-Host '  Администратор  admin@sportreg.local / Admin12345'
Write-Host '  Пользователи   ivan.petrov@mail.ru / Sport2026, olga.smirnova@mail.ru / Chess2026,'
Write-Host '                 maria.kuznetsova@mail.ru / Volley2026'
Write-Host ''
Write-Host 'Что показывать по каждой лабораторной — DEMO.md'

if ($Open) {
    Start-Process 'http://localhost:8091'
    Start-Process 'http://localhost:3001/d/sportreg-overview'
    Start-Process 'http://localhost:9091/targets'
}
