import { useCallback, useMemo, useState } from 'react'

import {
  changeApplicationStatus,
  fetchMyApplications,
  submitApplication,
} from '../api/applications-api'
import { ApplicationsTable } from '../components/elements/applications-table'
import { CompetitionCard } from '../components/elements/competition-card'
import { CompetitionFilters } from '../components/elements/competition-filters'
import { StatusBadge } from '../components/ui/status-badge'
import { useCompetitions } from '../hooks/requests/use-competitions'
import { useRequest } from '../hooks/requests/use-request'
import type { Application } from '../types/application'
import type { Competition } from '../types/competition'

const NO_APPLICATIONS: Application[] = []

const isActive = (a: Application) => a.status === 'pending' || a.status === 'approved'

// Страница пользователя (USER и ADMIN): каталог соревнований, подача и отмена заявок
export function UserPage() {
  const [query, setQuery] = useState('')
  const [sport, setSport] = useState('')
  const filter = useMemo(() => ({ q: query.trim(), sport: sport.trim() }), [query, sport])
  const competitions = useCompetitions(filter)
  const loadMine = useCallback(() => fetchMyApplications(), [])
  const mine = useRequest(loadMine, NO_APPLICATIONS)
  const [message, setMessage] = useState<{ error: boolean; text: string } | null>(null)

  const activeByCompetition = useMemo(
    () => new Map(mine.data.filter(isActive).map((a) => [a.competition_id, a])),
    [mine.data],
  )

  async function run(action: () => Promise<unknown>, success: string) {
    try {
      await action()
      setMessage({ error: false, text: success })
    } catch (err) {
      setMessage({ error: true, text: err instanceof Error ? err.message : 'Ошибка' })
    }
    competitions.reload()
    mine.reload()
  }

  function cardAction(c: Competition) {
    const active = activeByCompetition.get(c.id)
    if (active) return <StatusBadge status={active.status} />
    return (
      <button
        className="button"
        disabled={c.free_places === 0}
        onClick={() => run(() => submitApplication(c.id), `Заявка на «${c.title}» подана`)}
      >
        {c.free_places === 0 ? 'Мест нет' : 'Подать заявку'}
      </button>
    )
  }

  return (
    <section>
      <div className="page-head">
        <h1>Соревнования</h1>
      </div>
      {message && (
        <p className={message.error ? 'alert alert--error' : 'alert alert--info'}>{message.text}</p>
      )}
      <CompetitionFilters
        query={query}
        sport={sport}
        onQueryChange={setQuery}
        onSportChange={setSport}
      />
      {competitions.error && <p className="alert alert--error">{competitions.error}</p>}
      {competitions.loading ? (
        <p className="muted">Загрузка…</p>
      ) : (
        <div className="cards">
          {competitions.data.map((c) => (
            <CompetitionCard key={c.id} competition={c} action={cardAction(c)} />
          ))}
        </div>
      )}

      <h2 className="section-title">Мои заявки</h2>
      {mine.error && <p className="alert alert--error">{mine.error}</p>}
      <ApplicationsTable
        applications={mine.data}
        actions={(a) =>
          isActive(a) && (
            <button
              className="button button--small button--danger"
              onClick={() =>
                run(() => changeApplicationStatus(a.id, 'cancelled'), 'Заявка отменена')
              }
            >
              Отменить
            </button>
          )
        }
      />
    </section>
  )
}
