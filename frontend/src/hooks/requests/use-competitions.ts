import { useCallback } from 'react'

import { fetchCompetitions } from '../../api/competitions-api'
import type { Competition, CompetitionFilter } from '../../types/competition'
import { useRequest } from './use-request'

const EMPTY: Competition[] = []

// Каталог соревнований, перезагружается при смене фильтра
export function useCompetitions(filter: CompetitionFilter) {
  const load = useCallback(() => fetchCompetitions(filter), [filter])
  return useRequest(load, EMPTY)
}
