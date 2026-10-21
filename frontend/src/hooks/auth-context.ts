import { createContext } from 'react'

import type { User } from '../types/user'

export interface AuthState {
  user: User | null
  login: (email: string, password: string) => Promise<User>
  logout: () => void
}

export const AuthContext = createContext<AuthState | null>(null)
