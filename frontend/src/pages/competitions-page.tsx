import { useMemo, useState } from 'react'

import { createCompetition, deleteCompetition, updateCompetition } from '../api/competitions-api'
import { CompetitionForm } from '../components/elements/competition-form'
import { CompetitionTable } from '../components/elements/competition-table'
import { useCompetitions } from '../hooks/requests/use-competitions'
import type { Competition, CompetitionInput } from '../types/competition'

// Для PATCH оставляем только поля, которые пользователь действительно изменил
function changedFields(before: Competition, after: CompetitionInput): Partial<CompetitionInput> {
  return Object.fromEntries(
    Object.entries(after).filter(([key, value]) => before[key as keyof CompetitionInput] !== value),
  )
}

export function CompetitionsPage() {
  const [query, setQuery] = useState('')
  const [sport, setSport] = useState('')
  const filter = useMemo(() => ({ q: query.trim(), sport: sport.trim() }), [query, sport])
  const { competitions, loading, error, reload } = useCompetitions(filter)

  // undefined — форма закрыта, null — создание, Competition — изменение
  const [editing, setEditing] = useState<Competition | null | undefined>(undefined)
  const [actionError, setActionError] = useState<string | null>(null)

  async function handleSubmit(input: CompetitionInput) {
    if (editing) {
      await updateCompetition(editing.id, changedFields(editing, input))
    } else {
      await createCompetition(input)
    }
    setEditing(undefined)
    reload()
  }

  async function handleDelete(c: Competition) {
    if (!window.confirm(`Удалить «${c.title}»?`)) return
    try {
      await deleteCompetition(c.id)
      setActionError(null)
      reload()
    } catch (err) {
      setActionError(err instanceof Error ? err.message : 'Не удалось удалить')
    }
  }

  return (
    <section>
      <div className="page-head">
        <h1>Соревнования</h1>
        <button className="button" onClick={() => setEditing(null)}>
          Добавить соревнование
        </button>
      </div>

      {editing !== undefined && (
        <CompetitionForm
          key={editing?.id ?? 'new'}
          initial={editing ?? undefined}
          onSubmit={handleSubmit}
          onCancel={() => setEditing(undefined)}
        />
      )}

      <div className="filters">
        <input
          placeholder="Поиск по названию"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
        />
        <input placeholder="Вид спорта" value={sport} onChange={(e) => setSport(e.target.value)} />
      </div>

      {(error ?? actionError) && <p className="alert alert--error">{error ?? actionError}</p>}
      {loading ? (
        <p className="muted">Загрузка…</p>
      ) : (
        <CompetitionTable competitions={competitions} onEdit={setEditing} onDelete={handleDelete} />
      )}
    </section>
  )
}
