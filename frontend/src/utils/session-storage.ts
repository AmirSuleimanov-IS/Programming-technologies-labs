import type { Session } from '../types/user'

const KEY = 'sportreg.session'

// Сессия (JWT и пользователь) хранится в localStorage, чтобы пережить перезагрузку страницы.
// Риск: при XSS скрипт может прочитать токен. Для учебного проекта это принято осознанно (см. README).
export function loadSession(): Session | null {
  try {
    const session = JSON.parse(localStorage.getItem(KEY) ?? 'null') as Session | null
    if (session && new Date(session.expires_at) > new Date()) return session
  } catch {
    // повреждённая запись — считаем, что сессии нет
  }
  localStorage.removeItem(KEY)
  return null
}

export function saveSession(session: Session): void {
  localStorage.setItem(KEY, JSON.stringify(session))
}

export function clearSession(): void {
  localStorage.removeItem(KEY)
}
