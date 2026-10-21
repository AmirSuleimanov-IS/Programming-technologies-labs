import type { Competition } from '../../types/competition'
import { formatDate } from '../../utils/format-date'

interface CompetitionTableProps {
  competitions: Competition[]
  onEdit: (competition: Competition) => void
  onDelete: (competition: Competition) => void
}

export function CompetitionTable({ competitions, onEdit, onDelete }: CompetitionTableProps) {
  if (competitions.length === 0) {
    return <p className="empty">Соревнований не найдено</p>
  }
  return (
    <table className="table">
      <thead>
        <tr>
          <th>Название</th>
          <th>Вид спорта</th>
          <th>Дата</th>
          <th>Место</th>
          <th className="table__num">Мест</th>
          <th />
        </tr>
      </thead>
      <tbody>
        {competitions.map((c) => (
          <tr key={c.id}>
            <td>
              <strong>{c.title}</strong>
              {c.description && <div className="muted">{c.description}</div>}
            </td>
            <td>{c.sport}</td>
            <td className="nowrap">{formatDate(c.starts_on)}</td>
            <td>{c.location}</td>
            <td className="table__num">{c.capacity}</td>
            <td className="table__actions">
              <button className="button button--small" onClick={() => onEdit(c)}>
                Изменить
              </button>
              <button className="button button--small button--danger" onClick={() => onDelete(c)}>
                Удалить
              </button>
            </td>
          </tr>
        ))}
      </tbody>
    </table>
  )
}
