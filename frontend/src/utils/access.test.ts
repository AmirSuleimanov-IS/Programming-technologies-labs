import { describe, expect, it } from 'vitest'

import { PAGE_ROLES } from '../constants/routes'
import { canAccess } from './access'

// Матрица доступа к страницам: USER — одна защищённая страница, ADMIN — все, гость — ни одной
describe('canAccess', () => {
  it.each([
    ['гость', undefined, false, false],
    ['USER', 'USER', true, false],
    ['ADMIN', 'ADMIN', true, true],
  ] as const)('%s', (_, role, userPage, adminPage) => {
    expect(canAccess(role, PAGE_ROLES.userPage)).toBe(userPage)
    expect(canAccess(role, PAGE_ROLES.adminPage)).toBe(adminPage)
  })
})
