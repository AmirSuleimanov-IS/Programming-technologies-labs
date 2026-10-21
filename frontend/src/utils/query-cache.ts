interface Entry {
  value: unknown
  storedAt: number
  expiresAt: number
}

// Кэш ответов API в памяти вкладки по стратегии cache-aside с TTL. Живёт до перезагрузки
// страницы и не попадает в localStorage: данные не переживают сессию и не остаются на чужом компьютере.
// Каждая операция пишется в консоль браузера — так видно, откуда пришли данные.
export class QueryCache {
  private readonly entries = new Map<string, Entry>()
  private readonly now: () => number

  constructor(now: () => number = Date.now) {
    this.now = now
  }

  get<T>(key: string): { value: T; ageMs: number } | undefined {
    const entry = this.entries.get(key)
    if (!entry) {
      log('miss', key)
      return undefined
    }
    if (entry.expiresAt <= this.now()) {
      this.entries.delete(key)
      log('miss (TTL истёк)', key)
      return undefined
    }
    const ageMs = this.now() - entry.storedAt
    log(`hit, возраст ${Math.round(ageMs / 1000)} с`, key)
    return { value: entry.value as T, ageMs }
  }

  set(key: string, value: unknown, ttlMs: number): void {
    const now = this.now()
    this.entries.set(key, { value, storedAt: now, expiresAt: now + ttlMs })
    log(`set, TTL ${Math.round(ttlMs / 1000)} с`, key)
  }

  // Удаляет все записи, ключ которых начинается с prefix
  invalidate(prefix: string): void {
    for (const key of this.entries.keys()) {
      if (key.startsWith(prefix)) this.entries.delete(key)
    }
    log('invalidate', prefix + '*')
  }

  clear(): void {
    this.entries.clear()
    log('clear', '*')
  }
}

function log(result: string, key: string): void {
  console.info(`[cache] ${result}: ${key}`)
}

export const queryCache = new QueryCache()
