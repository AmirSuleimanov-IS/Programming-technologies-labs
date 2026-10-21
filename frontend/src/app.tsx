import { BrowserRouter, Navigate, Route, Routes } from 'react-router'

import { AppLayout } from './components/elements/app-layout'
import { AuthProvider } from './components/elements/auth-provider'
import { ProtectedRoute } from './components/elements/protected-route'
import { PAGE_ROLES, ROUTES } from './constants/routes'
import { AdminPage } from './pages/admin-page'
import { ForbiddenPage } from './pages/forbidden-page'
import { LoginPage } from './pages/login-page'
import { NotFoundPage } from './pages/not-found-page'
import { RegisterPage } from './pages/register-page'
import { UserPage } from './pages/user-page'

export function App() {
  return (
    <AuthProvider>
      <BrowserRouter>
        <Routes>
          <Route element={<AppLayout />}>
            <Route index element={<Navigate to={ROUTES.userPage} replace />} />
            <Route path={ROUTES.login} element={<LoginPage />} />
            <Route path={ROUTES.register} element={<RegisterPage />} />
            <Route path={ROUTES.forbidden} element={<ForbiddenPage />} />
            <Route element={<ProtectedRoute roles={PAGE_ROLES.userPage} />}>
              <Route path={ROUTES.userPage} element={<UserPage />} />
            </Route>
            <Route element={<ProtectedRoute roles={PAGE_ROLES.adminPage} />}>
              <Route path={ROUTES.adminPage} element={<AdminPage />} />
            </Route>
            <Route path="*" element={<NotFoundPage />} />
          </Route>
        </Routes>
      </BrowserRouter>
    </AuthProvider>
  )
}
