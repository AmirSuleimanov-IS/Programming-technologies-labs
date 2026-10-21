import type { Role } from '../types/user'

export const ROUTES = {
  login: '/login',
  register: '/register',
  userPage: '/competitions',
  adminPage: '/admin',
  forbidden: '/forbidden',
} as const

// Кто может открыть защищённые страницы. USER видит одну страницу, ADMIN — все.
// Это только удобство интерфейса: права проверяет бэкенд.
export const PAGE_ROLES = {
  userPage: ['USER', 'ADMIN'],
  adminPage: ['ADMIN'],
} as const satisfies Record<string, readonly Role[]>
