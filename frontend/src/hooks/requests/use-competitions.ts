import { useCallback, useEffect, useState } from 'react'

import { fetchCompetitions } from '../../api/competitions-api'
import type { Competition, CompetitionFilter } from '../../types/competition'

// Загружает каталог соревнований и перезагружает его при смене фильтра
export function useCompetitions(filter: CompetitionFilter) {
  const [competitions, setCompetitions] = useState<Competition[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [version, setVersion] = useState(0)

  useEffect(() => {
    let cancelled = false
    fetchCompetitions(filter)
      .then((list) => {
        if (!cancelled) {
          setCompetitions(list)
          setError(null)
        }
      })
      .catch((err: Error) => {
        if (!cancelled) setError(err.message)
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [filter, version])

  const reload = useCallback(() => {
    setLoading(true)
    setVersion((v) => v + 1)
  }, [])

  return { competitions, loading, error, reload }
}
