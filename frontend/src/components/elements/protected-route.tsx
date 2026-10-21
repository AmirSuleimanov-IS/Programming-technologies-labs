import { Navigate, Outlet, useLocation } from 'react-router'

import { ROUTES } from '../../constants/routes'
import { useAuth } from '../../hooks/use-auth'
import type { Role } from '../../types/user'
import { canAccess } from '../../utils/access'

// Пропускает на вложенные маршруты только вошедших пользователей с подходящей ролью.
// Гость попадает на страницу входа, пользователь без прав — на страницу 403.
export function ProtectedRoute({ roles }: { roles: readonly Role[] }) {
  const { user } = useAuth()
  const location = useLocation()
  if (!user) {
    return <Navigate to={ROUTES.login} replace state={{ from: location.pathname }} />
  }
  if (!canAccess(user.role, roles)) {
    return <Navigate to={ROUTES.forbidden} replace />
  }
  return <Outlet />
}
