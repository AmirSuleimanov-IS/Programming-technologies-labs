import { Link } from 'react-router'

import { ROUTES } from '../constants/routes'

export function ForbiddenPage() {
  return (
    <section className="panel center">
      <h1>403 — доступ запрещён</h1>
      <p className="muted">Эта страница доступна только администратору.</p>
      <Link className="button" to={ROUTES.userPage}>
        К соревнованиям
      </Link>
    </section>
  )
}
