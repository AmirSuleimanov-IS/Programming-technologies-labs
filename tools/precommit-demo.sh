#!/usr/bin/env bash
# Показ pre-commit для лабы 1: на временной ветке по очереди делаются коммиты с нарушениями,
# и каждый должен быть остановлен. В конце ветка и временные файлы удаляются.
# Запуск из Git Bash в папке technologies:  bash tools/precommit-demo.sh
# (в PowerShell:  & "C:\Program Files\Git\bin\bash.exe" tools/precommit-demo.sh)
set -uo pipefail
cd "$(dirname "$0")/.."

if [ -n "$(git status --porcelain)" ]; then
  echo "В рабочей копии есть незакоммиченные изменения — сначала закоммитьте или уберите их."
  exit 1
fi
start=$(git branch --show-current)
created=(
  backend/internal/database/dump.bin
  docs/limits.md
  backend/internal/competition/capacity.go
  frontend/src/utils/label.ts
  frontend/src/components/elements/CompetitionCard.tsx
  backend/internal/competition/competitionCache.go
)
cleanup() {
  git reset -q --hard
  rm -f "${created[@]}"
  git switch -q "$start"
  git branch -q -D demo-precommit 2> /dev/null
  echo "Временная ветка удалена, вы снова на ветке $start."
}
trap cleanup EXIT
git switch -q -c demo-precommit

pause() {
  if [ -t 0 ]; then read -r -p $'\n— Enter: следующий случай —\n'; fi
}
attempt() {
  echo "\$ git commit -m \"$1\""
  git commit -q -m "$1" 2>&1 | grep -v "Skipped"
  echo "[код возврата git commit: ${PIPESTATUS[0]}]"
  git reset -q --hard
  rm -f "${created[@]}"
}
title() {
  echo
  echo "================ $1 ================"
}

title "1. Файл больше 500 КБ"
head -c 700000 /dev/urandom > backend/internal/database/dump.bin
git add backend/internal/database/dump.bin
attempt "добавить дамп"
pause

title "2. Маркеры неразрешённого конфликта"
printf '# Заметки\n\n<<<<<<< HEAD\nЛимит мест: 10\n=======\nЛимит мест: 20\n>>>>>>> feature\n' > docs/limits.md
git add docs/limits.md
attempt "заметки о лимитах"
pause

title "3. Неотформатированный код и неиспользуемая функция"
printf 'package competition\n\nfunc   isFull(approved,capacity int) bool{\nreturn approved>=capacity}\n' > backend/internal/competition/capacity.go
printf 'import { formatDate } from "./format-date"\nexport function label(d:string){return "Старт: "+formatDate(d)}\n' > frontend/src/utils/label.ts
git add backend/internal/competition/capacity.go frontend/src/utils/label.ts
attempt "форматирование"
pause

title "4. Неверные имена файлов (стандарт: Go — snake_case, фронтенд — kebab-case)"
printf 'export function CompetitionCard() {\n  return null\n}\n' > frontend/src/components/elements/CompetitionCard.tsx
printf 'package competition\n' > backend/internal/competition/competitionCache.go
git add frontend/src/components/elements/CompetitionCard.tsx backend/internal/competition/competitionCache.go
attempt "карточка соревнования"
pause

title "5. Автоисправление: prettier правит файл и останавливает коммит, разработчик проверяет и повторяет"
sed -i "s/  const res = await request<Competition>('POST', '\/competitions', input)/  const res = await request<Competition>(\"POST\", '\/competitions', input);/" frontend/src/api/competitions-api.ts
sed -i "s/^export async function createCompetition(/\/\/ POST: новое соревнование\nexport async function createCompetition(/" frontend/src/api/competitions-api.ts
git add frontend/src/api/competitions-api.ts
echo '$ git commit -m "комментарий к запросу создания"'
git commit -q -m "комментарий к запросу создания" 2>&1 | grep -v "Skipped"
echo "[код возврата git commit: ${PIPESTATUS[0]}]"
echo
echo '$ git diff   # что исправил prettier'
git --no-pager diff
pause
echo '$ git add frontend/src/api/competitions-api.ts && git commit -m "комментарий к запросу создания"'
git add frontend/src/api/competitions-api.ts
git commit -q -m "комментарий к запросу создания" 2>&1 | grep -v "Skipped"
echo "[код возврата git commit: ${PIPESTATUS[0]}]"
