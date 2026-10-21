import { beforeEach, describe, expect, it, vi } from 'vitest'

import { queryCache } from '../utils/query-cache'
import { createCompetition, fetchCompetitions } from './competitions-api'
import { request } from './http-client'

vi.mock('./http-client', () => ({ request: vi.fn() }))
const requestMock = vi.mocked(request)

const filter = { q: '', sport: '' }

describe('fetchCompetitions: кэш каталога в браузере', () => {
  beforeEach(() => {
    vi.spyOn(console, 'info').mockImplementation(() => {})
    queryCache.clear()
    requestMock.mockReset()
    requestMock.mockResolvedValue({ data: [{ id: 1 }], meta: { cache: 'miss' } })
  })

  it('первый раз — сервер (источник из meta.cache), второй — кэш браузера', async () => {
    expect(await fetchCompetitions(filter)).toMatchObject({ source: 'miss' })
    expect(await fetchCompetitions(filter)).toMatchObject({ source: 'browser' })
    expect(requestMock).toHaveBeenCalledTimes(1)
  })

  it('фильтр в другом регистре — тот же ключ', async () => {
    await fetchCompetitions({ q: 'Кубок', sport: '' })
    expect(await fetchCompetitions({ q: 'кубок', sport: '' })).toMatchObject({ source: 'browser' })
  })

  it('одновременные одинаковые запросы объединяются в один', async () => {
    await Promise.all([
      fetchCompetitions(filter),
      fetchCompetitions(filter),
      fetchCompetitions(filter),
    ])
    expect(requestMock).toHaveBeenCalledTimes(1)
  })

  it('создание соревнования сбрасывает кэш каталога', async () => {
    await fetchCompetitions(filter)
    requestMock.mockResolvedValueOnce({ data: { id: 2 } })
    await createCompetition({
      title: 'Кубок',
      sport: 'Футбол',
      location: 'Норильск',
      starts_on: '2026-12-01',
      capacity: 10,
      description: '',
    })
    expect(await fetchCompetitions(filter)).toMatchObject({ source: 'miss' })
    expect(requestMock).toHaveBeenCalledTimes(3)
  })

  it('ошибка сервера не кэшируется', async () => {
    requestMock.mockRejectedValueOnce(new Error('Сервер недоступен'))
    await expect(fetchCompetitions(filter)).rejects.toThrow('Сервер недоступен')
    expect(await fetchCompetitions(filter)).toMatchObject({ source: 'miss' })
  })
})
