import { Link } from 'react-router'

export function NotFoundPage() {
  return (
    <section className="panel center">
      <h1>Страница не найдена</h1>
      <p className="muted">Проверьте адрес или вернитесь в каталог.</p>
      <Link className="button" to="/">
        К соревнованиям
      </Link>
    </section>
  )
}
