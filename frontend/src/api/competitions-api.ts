import type { Competition, CompetitionFilter, CompetitionInput } from '../types/competition'
import type { Cached, ServerCacheMeta } from '../types/cache'
import { queryCache } from '../utils/query-cache'
import { request } from './http-client'

const CACHE_PREFIX = 'competitions'
const CACHE_TTL_MS = Number(import.meta.env.VITE_CACHE_TTL_SECONDS ?? 30) * 1000

// Запросы, которые уже отправлены, но ещё не завершились: одинаковые промахи
// не порождают повторных запросов к серверу (как singleflight на бэкенде)
const inflight = new Map<string, Promise<Cached<Competition[]>>>()

// Каталог: сначала кэш браузера, при промахе — запрос к серверу. Сервер сообщает в meta.cache,
// ответил он из Redis (hit) или из базы данных (miss, bypass).
export async function fetchCompetitions(filter: CompetitionFilter): Promise<Cached<Competition[]>> {
  const key = `${CACHE_PREFIX}?q=${filter.q.toLowerCase()}&sport=${filter.sport.toLowerCase()}`
  const cached = queryCache.get<Competition[]>(key)
  if (cached) {
    return { data: cached.value, source: 'browser', ageMs: cached.ageMs }
  }
  let pending = inflight.get(key)
  if (!pending) {
    pending = loadFromServer(key, filter).finally(() => inflight.delete(key))
    inflight.set(key, pending)
  }
  return pending
}

async function loadFromServer(
  key: string,
  filter: CompetitionFilter,
): Promise<Cached<Competition[]>> {
  const params = new URLSearchParams()
  if (filter.q) params.set('q', filter.q)
  if (filter.sport) params.set('sport', filter.sport)
  const query = params.toString()
  const res = await request<Competition[], ServerCacheMeta>(
    'GET',
    '/competitions' + (query ? `?${query}` : ''),
  )
  queryCache.set(key, res.data, CACHE_TTL_MS)
  return { data: res.data, source: res.meta?.cache ?? 'server', ageMs: 0 }
}

// После любого изменения каталога или заявок (меняются свободные места) кэш каталога сбрасывается
export function invalidateCompetitions(): void {
  queryCache.invalidate(CACHE_PREFIX)
}

export async function createCompetition(input: CompetitionInput): Promise<Competition> {
  const res = await request<Competition>('POST', '/competitions', input)
  invalidateCompetitions()
  return res.data
}

// PATCH: отправляются только изменённые поля
export async function updateCompetition(
  id: number,
  changes: Partial<CompetitionInput>,
): Promise<Competition> {
  const res = await request<Competition>('PATCH', `/competitions/${id}`, changes)
  invalidateCompetitions()
  return res.data
}

export async function deleteCompetition(id: number): Promise<void> {
  await request<void>('DELETE', `/competitions/${id}`)
  invalidateCompetitions()
}
