import { useContext } from 'react'

import { AuthContext, type AuthState } from './auth-context'

export function useAuth(): AuthState {
  const auth = useContext(AuthContext)
  if (!auth) throw new Error('useAuth вызван вне AuthProvider')
  return auth
}
