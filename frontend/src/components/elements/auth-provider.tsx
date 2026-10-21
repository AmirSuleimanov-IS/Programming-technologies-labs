import { useCallback, useEffect, useMemo, useState, type ReactNode } from 'react'

import { login as loginRequest } from '../../api/auth-api'
import { setAuthToken, setUnauthorizedHandler } from '../../api/http-client'
import { AuthContext } from '../../hooks/auth-context'
import type { Session, User } from '../../types/user'
import { queryCache } from '../../utils/query-cache'
import { clearSession, loadSession, saveSession } from '../../utils/session-storage'

function restoreSession(): Session | null {
  const session = loadSession()
  setAuthToken(session?.token ?? null) // до первых запросов дочерних компонентов
  return session
}

// Хранит сессию: вход, выход, восстановление после перезагрузки страницы
export function AuthProvider({ children }: { children: ReactNode }) {
  const [session, setSession] = useState<Session | null>(restoreSession)

  // Кэш браузера очищается при входе и выходе: следующий пользователь за этим компьютером
  // не должен видеть данные, загруженные предыдущим
  const logout = useCallback(() => {
    queryCache.clear()
    clearSession()
    setAuthToken(null)
    setSession(null)
  }, [])

  const login = useCallback(async (email: string, password: string): Promise<User> => {
    const next = await loginRequest(email, password)
    queryCache.clear()
    saveSession(next)
    setAuthToken(next.token)
    setSession(next)
    return next.user
  }, [])

  // Сервер ответил 401 на запрос с токеном — токен истёк, выходим
  useEffect(() => setUnauthorizedHandler(logout), [logout])

  const value = useMemo(
    () => ({ user: session?.user ?? null, login, logout }),
    [session, login, logout],
  )
  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}
