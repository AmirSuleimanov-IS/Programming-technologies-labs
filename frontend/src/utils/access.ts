import type { Role } from '../types/user'

// Может ли пользователь с ролью role открыть страницу, доступную ролям allowed
export function canAccess(role: Role | undefined, allowed: readonly Role[]): boolean {
  return role !== undefined && allowed.includes(role)
}
