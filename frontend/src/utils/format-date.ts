const DATE_FORMAT = new Intl.DateTimeFormat('ru-RU', {
  day: 'numeric',
  month: 'long',
  year: 'numeric',
})

// "2026-11-04" → "4 ноября 2026 г."
export function formatDate(isoDate: string): string {
  const [year, month, day] = isoDate.split('-').map(Number)
  return DATE_FORMAT.format(new Date(year, month - 1, day))
}

// Сегодняшняя дата в формате ГГГГ-ММ-ДД по местному времени (для min у <input type="date">)
export function todayIso(): string {
  const now = new Date()
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${now.getFullYear()}-${pad(now.getMonth() + 1)}-${pad(now.getDate())}`
}
