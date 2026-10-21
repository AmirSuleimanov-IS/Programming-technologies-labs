import type { Session, User } from '../types/user'
import { request } from './http-client'

export async function login(email: string, password: string): Promise<Session> {
  const res = await request<Session>('POST', '/auth/login', { email, password })
  return res.data
}

export async function register(email: string, fullName: string, password: string): Promise<User> {
  const res = await request<User>('POST', '/auth/register', {
    email,
    full_name: fullName,
    password,
  })
  return res.data
}
