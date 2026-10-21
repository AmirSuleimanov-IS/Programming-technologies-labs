import type { DataSource } from '../../types/cache'

const LABELS: Record<DataSource, string> = {
  browser: 'кэш браузера',
  hit: 'сервер, кэш Redis (HIT)',
  miss: 'сервер, база данных (MISS)',
  bypass: 'сервер, база данных — Redis недоступен (BYPASS)',
  server: 'сервер',
}

interface SourceBadgeProps {
  source: DataSource
  ageMs: number
  onRefresh: () => void
}

// Показывает, откуда пришли данные, и позволяет запросить их у сервера в обход кэша браузера
export function SourceBadge({ source, ageMs, onRefresh }: SourceBadgeProps) {
  return (
    <div className={`source source--${source}`}>
      <span>
        Источник данных: <strong>{LABELS[source]}</strong>
        {source === 'browser' && `, сохранено ${Math.round(ageMs / 1000)} с назад`}
      </span>
      <button className="button button--small button--ghost" onClick={onRefresh}>
        Обновить с сервера
      </button>
    </div>
  )
}
