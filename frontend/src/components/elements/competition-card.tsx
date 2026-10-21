import type { ReactNode } from 'react'

import type { Competition } from '../../types/competition'
import { formatDate } from '../../utils/format-date'

interface CompetitionCardProps {
  competition: Competition
  action: ReactNode
}

export function CompetitionCard({ competition: c, action }: CompetitionCardProps) {
  const taken = c.capacity - c.free_places
  return (
    <article className="card">
      <div className="card__sport">{c.sport}</div>
      <h3 className="card__title">{c.title}</h3>
      <div className="muted">
        {formatDate(c.starts_on)} · {c.location}
      </div>
      {c.description && <p className="card__text">{c.description}</p>}
      <div className="places">
        <div className="places__bar">
          <div className="places__fill" style={{ width: `${(taken / c.capacity) * 100}%` }} />
        </div>
        <span
          className={c.free_places === 0 ? 'places__label places__label--full' : 'places__label'}
        >
          Свободно {c.free_places} из {c.capacity}
        </span>
      </div>
      <div className="card__action">{action}</div>
    </article>
  )
}
