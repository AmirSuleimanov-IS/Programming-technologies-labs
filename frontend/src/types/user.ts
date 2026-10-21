export type Role = 'ADMIN' | 'USER'

export interface User {
  id: number
  email: string
  full_name: string
  role: Role
  created_at: string
}

export interface Session {
  token: string
  expires_at: string
  user: User
}
