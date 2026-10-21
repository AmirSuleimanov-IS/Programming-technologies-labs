#!/bin/sh
# Сборка отчёта: sh tools/report/make.sh <N>
# reports/lab_N/spec.py → reports/lab_N/Отчёт_...docx; Word обновляет оглавление,
# проверочный PDF и PNG страниц кладутся в tools/raw/preview
set -e
N=$1
ROOT=$(cd "$(dirname "$0")/../.." && pwd)
PY="$ROOT/tools/.venv/Scripts/python"
OUT="$ROOT/reports/lab_$N/Отчёт_Лабораторная_работа_№${N}_ИС23_Сулейманов_А_Ш_.docx"
PREV="$ROOT/tools/raw/preview"
mkdir -p "$PREV"
"$PY" -I "$ROOT/tools/report/build_report.py" "$ROOT/reports/lab_$N/spec.py" "$OUT"
powershell -NoProfile -ExecutionPolicy Bypass -File "$(cygpath -w "$ROOT/tools/report/finalize.ps1")" -In "$(cygpath -w "$OUT")" -Pdf "$(cygpath -w "$PREV")\lab_$N.pdf"
rm -f "$PREV"/lab_${N}_p_*.png
"$PY" -I "$ROOT/tools/report/pdf2png.py" "$PREV/lab_$N.pdf" "$PREV/lab_${N}_p"
