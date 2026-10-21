import { useCallback } from 'react'

import { fetchCompetitions, invalidateCompetitions } from '../../api/competitions-api'
import type { Cached } from '../../types/cache'
import type { Competition, CompetitionFilter } from '../../types/competition'
import { useRequest } from './use-request'

const EMPTY: Cached<Competition[]> = { data: [], source: 'server', ageMs: 0 }

// Каталог соревнований через кэш браузера, перезагружается при смене фильтра.
// refresh сбрасывает кэш браузера и запрашивает данные у сервера.
export function useCompetitions(filter: CompetitionFilter) {
  const load = useCallback(() => fetchCompetitions(filter), [filter])
  const { data, loading, error, reload } = useRequest(load, EMPTY)

  const refresh = useCallback(() => {
    invalidateCompetitions()
    reload()
  }, [reload])

  return {
    data: data.data,
    source: data.source,
    ageMs: data.ageMs,
    loading,
    error,
    reload,
    refresh,
  }
}
