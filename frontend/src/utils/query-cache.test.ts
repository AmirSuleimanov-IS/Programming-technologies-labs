import { beforeEach, describe, expect, it, vi } from 'vitest'

import { QueryCache } from './query-cache'

describe('QueryCache', () => {
  let now: number
  let cache: QueryCache

  beforeEach(() => {
    vi.spyOn(console, 'info').mockImplementation(() => {}) // журнал кэша в консоли тестам не нужен
    now = 1_000_000
    cache = new QueryCache(() => now)
  })

  it('промах, затем попадание с возрастом записи', () => {
    expect(cache.get('competitions?q=')).toBeUndefined()
    cache.set('competitions?q=', ['кубок'], 30_000)
    now += 12_000
    expect(cache.get('competitions?q=')).toEqual({ value: ['кубок'], ageMs: 12_000 })
  })

  it('граница TTL: за 1 мс до истечения — попадание, в момент истечения — промах', () => {
    cache.set('k', 1, 30_000)
    now += 29_999
    expect(cache.get('k')).toBeDefined()
    now += 1
    expect(cache.get('k')).toBeUndefined()
    // истёкшая запись удалена и после продления времени не «воскресает»
    now -= 10_000
    expect(cache.get('k')).toBeUndefined()
  })

  it('инвалидация по префиксу удаляет только ключи этого префикса', () => {
    cache.set('competitions?q=', 1, 30_000)
    cache.set('competitions?q=кубок', 2, 30_000)
    cache.set('applications/my', 3, 30_000)
    cache.invalidate('competitions')
    expect(cache.get('competitions?q=')).toBeUndefined()
    expect(cache.get('competitions?q=кубок')).toBeUndefined()
    expect(cache.get('applications/my')?.value).toBe(3)
  })

  it('clear удаляет всё (выход из системы)', () => {
    cache.set('a', 1, 30_000)
    cache.set('b', 2, 30_000)
    cache.clear()
    expect(cache.get('a')).toBeUndefined()
    expect(cache.get('b')).toBeUndefined()
  })
})
