import { useCallback, useState } from 'react'

import { changeApplicationStatus, fetchApplications } from '../api/applications-api'
import { fetchUsers } from '../api/users-api'
import { ApplicationsTable } from '../components/elements/applications-table'
import { CompetitionsAdmin } from '../components/elements/competitions-admin'
import { STATUS_TITLES } from '../constants/application-status'
import { useRequest } from '../hooks/requests/use-request'
import type { Application, ApplicationStatus } from '../types/application'
import type { User } from '../types/user'

type Tab = 'applications' | 'competitions' | 'users'

const TABS: Record<Tab, string> = {
  applications: 'Заявки',
  competitions: 'Соревнования',
  users: 'Пользователи',
}

// Страница администратора (только ADMIN)
export function AdminPage() {
  const [tab, setTab] = useState<Tab>('applications')
  return (
    <section>
      <div className="page-head">
        <h1>Администрирование</h1>
      </div>
      <div className="tabs">
        {(Object.keys(TABS) as Tab[]).map((t) => (
          <button
            key={t}
            className={t === tab ? 'tab tab--active' : 'tab'}
            onClick={() => setTab(t)}
          >
            {TABS[t]}
          </button>
        ))}
      </div>
      {tab === 'applications' && <ApplicationsAdmin />}
      {tab === 'competitions' && <CompetitionsAdmin />}
      {tab === 'users' && <UsersAdmin />}
    </section>
  )
}

const NO_APPLICATIONS: Application[] = []
const NO_USERS: User[] = []

function ApplicationsAdmin() {
  const [status, setStatus] = useState<ApplicationStatus | ''>('pending')
  const load = useCallback(() => fetchApplications(status), [status])
  const { data, loading, error, reload } = useRequest(load, NO_APPLICATIONS)
  const [actionError, setActionError] = useState<string | null>(null)

  async function decide(a: Application, next: 'approved' | 'rejected') {
    try {
      await changeApplicationStatus(a.id, next)
      setActionError(null)
    } catch (err) {
      setActionError(err instanceof Error ? err.message : 'Ошибка')
    }
    reload()
  }

  return (
    <>
      <div className="filters">
        <select
          value={status}
          onChange={(e) => setStatus(e.target.value as ApplicationStatus | '')}
        >
          <option value="">Все статусы</option>
          {(Object.keys(STATUS_TITLES) as ApplicationStatus[]).map((s) => (
            <option key={s} value={s}>
              {STATUS_TITLES[s]}
            </option>
          ))}
        </select>
      </div>
      {(error ?? actionError) && <p className="alert alert--error">{error ?? actionError}</p>}
      {loading ? (
        <p className="muted">Загрузка…</p>
      ) : (
        <ApplicationsTable
          applications={data}
          showUser
          actions={(a) =>
            a.status === 'pending' && (
              <>
                <button className="button button--small" onClick={() => decide(a, 'approved')}>
                  Одобрить
                </button>
                <button
                  className="button button--small button--danger"
                  onClick={() => decide(a, 'rejected')}
                >
                  Отклонить
                </button>
              </>
            )
          }
        />
      )}
    </>
  )
}

function UsersAdmin() {
  const { data, loading, error } = useRequest(fetchUsers, NO_USERS)
  if (loading) return <p className="muted">Загрузка…</p>
  if (error) return <p className="alert alert--error">{error}</p>
  return (
    <table className="table">
      <thead>
        <tr>
          <th>ID</th>
          <th>Имя</th>
          <th>Email</th>
          <th>Роль</th>
          <th>Зарегистрирован</th>
        </tr>
      </thead>
      <tbody>
        {data.map((u) => (
          <tr key={u.id}>
            <td>{u.id}</td>
            <td>{u.full_name}</td>
            <td>{u.email}</td>
            <td>
              <span className={u.role === 'ADMIN' ? 'role role--admin' : 'role'}>{u.role}</span>
            </td>
            <td className="muted">{new Date(u.created_at).toLocaleString('ru-RU')}</td>
          </tr>
        ))}
      </tbody>
    </table>
  )
}
