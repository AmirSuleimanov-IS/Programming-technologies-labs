// Откуда пришли данные: кэш браузера, Redis на сервере (hit), база данных после промаха (miss)
// или база данных при недоступном Redis (bypass)
export type DataSource = 'browser' | 'hit' | 'miss' | 'bypass' | 'server'

export interface Cached<T> {
  data: T
  source: DataSource
  ageMs: number // возраст записи в кэше браузера
}

export interface ServerCacheMeta {
  count?: number
  cache?: 'hit' | 'miss' | 'bypass'
}
