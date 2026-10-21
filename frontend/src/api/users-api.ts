import type { User } from '../types/user'
import { request } from './http-client'

export async function fetchUsers(): Promise<User[]> {
  const res = await request<User[]>('GET', '/users')
  return res.data
}
