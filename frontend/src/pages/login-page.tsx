import { useState, type FormEvent } from 'react'
import { Link, Navigate, useLocation, useNavigate } from 'react-router'

import { FormField } from '../components/ui/form-field'
import { ROUTES } from '../constants/routes'
import { useAuth } from '../hooks/use-auth'

export function LoginPage() {
  const { user, login } = useAuth()
  const navigate = useNavigate()
  const location = useLocation()
  const from = (location.state as { from?: string } | null)?.from ?? ROUTES.userPage
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [sending, setSending] = useState(false)

  if (user) return <Navigate to={from} replace />

  async function handleSubmit(event: FormEvent) {
    event.preventDefault()
    setSending(true)
    setError(null)
    try {
      await login(email, password)
      navigate(from, { replace: true })
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Не удалось войти')
      setSending(false)
    }
  }

  return (
    <form className="panel form auth-form" onSubmit={handleSubmit}>
      <h1>Вход</h1>
      {error && <p className="alert alert--error">{error}</p>}
      <FormField label="Email">
        <input
          type="email"
          autoComplete="username"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
        />
      </FormField>
      <FormField label="Пароль">
        <input
          type="password"
          autoComplete="current-password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
        />
      </FormField>
      <button type="submit" className="button" disabled={sending}>
        {sending ? 'Вход…' : 'Войти'}
      </button>
      <p className="muted">
        Нет учётной записи? <Link to={ROUTES.register}>Зарегистрируйтесь</Link>
      </p>
    </form>
  )
}
