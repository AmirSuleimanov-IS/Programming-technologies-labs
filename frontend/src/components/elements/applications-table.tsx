import type { ReactNode } from 'react'

import type { Application } from '../../types/application'
import { formatDate } from '../../utils/format-date'
import { StatusBadge } from '../ui/status-badge'

interface ApplicationsTableProps {
  applications: Application[]
  showUser?: boolean
  actions: (application: Application) => ReactNode
}

export function ApplicationsTable({
  applications,
  showUser = false,
  actions,
}: ApplicationsTableProps) {
  if (applications.length === 0) {
    return <p className="empty">Заявок нет</p>
  }
  return (
    <table className="table">
      <thead>
        <tr>
          <th>Соревнование</th>
          <th>Дата</th>
          {showUser && <th>Пользователь</th>}
          <th>Статус</th>
          <th>Подана</th>
          <th />
        </tr>
      </thead>
      <tbody>
        {applications.map((a) => (
          <tr key={a.id}>
            <td>
              <strong>{a.competition_title}</strong>
            </td>
            <td className="nowrap">{formatDate(a.starts_on)}</td>
            {showUser && (
              <td>
                {a.user_full_name}
                <div className="muted">{a.user_email}</div>
              </td>
            )}
            <td>
              <StatusBadge status={a.status} />
            </td>
            <td className="nowrap muted">{new Date(a.created_at).toLocaleString('ru-RU')}</td>
            <td>
              <div className="table__actions">{actions(a)}</div>
            </td>
          </tr>
        ))}
      </tbody>
    </table>
  )
}
