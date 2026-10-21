import { useState, type FormEvent } from 'react'
import { Link, Navigate, useNavigate } from 'react-router'

import { register } from '../api/auth-api'
import { ApiError } from '../api/http-client'
import { FormField } from '../components/ui/form-field'
import { ROUTES } from '../constants/routes'
import { useAuth } from '../hooks/use-auth'

export function RegisterPage() {
  const { user, login } = useAuth()
  const navigate = useNavigate()
  const [email, setEmail] = useState('')
  const [fullName, setFullName] = useState('')
  const [password, setPassword] = useState('')
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({})
  const [error, setError] = useState<string | null>(null)
  const [sending, setSending] = useState(false)

  if (user) return <Navigate to={ROUTES.userPage} replace />

  async function handleSubmit(event: FormEvent) {
    event.preventDefault()
    setSending(true)
    setError(null)
    setFieldErrors({})
    try {
      await register(email, fullName, password)
      await login(email, password)
      navigate(ROUTES.userPage, { replace: true })
    } catch (err) {
      if (err instanceof ApiError && err.details.length > 0) {
        setFieldErrors(Object.fromEntries(err.details.map((d) => [d.field, d.message])))
      } else {
        setError(err instanceof Error ? err.message : 'Не удалось зарегистрироваться')
      }
      setSending(false)
    }
  }

  return (
    <form className="panel form auth-form" onSubmit={handleSubmit} noValidate>
      <h1>Регистрация</h1>
      {error && <p className="alert alert--error">{error}</p>}
      <FormField label="Email" error={fieldErrors.email}>
        <input type="email" value={email} onChange={(e) => setEmail(e.target.value)} />
      </FormField>
      <FormField label="Имя и фамилия" error={fieldErrors.full_name}>
        <input value={fullName} onChange={(e) => setFullName(e.target.value)} />
      </FormField>
      <FormField label="Пароль (от 8 символов, буквы и цифры)" error={fieldErrors.password}>
        <input
          type="password"
          autoComplete="new-password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
        />
      </FormField>
      <button type="submit" className="button" disabled={sending}>
        {sending ? 'Регистрация…' : 'Зарегистрироваться'}
      </button>
      <p className="muted">
        Уже есть учётная запись? <Link to={ROUTES.login}>Войдите</Link>
      </p>
    </form>
  )
}
