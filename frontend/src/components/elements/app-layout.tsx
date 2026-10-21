import { NavLink, Outlet, useNavigate } from 'react-router'

import { PAGE_ROLES, ROUTES } from '../../constants/routes'
import { useAuth } from '../../hooks/use-auth'
import { canAccess } from '../../utils/access'

const APP_TITLE = import.meta.env.VITE_APP_TITLE ?? 'SportReg'

export function AppLayout() {
  const { user, logout } = useAuth()
  const navigate = useNavigate()

  function handleLogout() {
    logout()
    navigate(ROUTES.login)
  }

  return (
    <>
      <header className="header">
        <div className="container header__inner">
          <NavLink to="/" className="logo">
            <img src="/favicon.svg" alt="" width={28} height={28} />
            {APP_TITLE}
          </NavLink>
          <nav className="nav">
            {canAccess(user?.role, PAGE_ROLES.userPage) && (
              <NavLink to={ROUTES.userPage}>Соревнования</NavLink>
            )}
            {canAccess(user?.role, PAGE_ROLES.adminPage) && (
              <NavLink to={ROUTES.adminPage}>Администрирование</NavLink>
            )}
          </nav>
          <div className="account">
            {user ? (
              <>
                <span className="account__name">{user.full_name}</span>
                <span className={user.role === 'ADMIN' ? 'role role--admin' : 'role'}>
                  {user.role}
                </span>
                <button className="button button--small button--light" onClick={handleLogout}>
                  Выйти
                </button>
              </>
            ) : (
              <>
                <NavLink to={ROUTES.login}>Вход</NavLink>
                <NavLink to={ROUTES.register}>Регистрация</NavLink>
              </>
            )}
          </div>
        </div>
      </header>
      <main className="container main">
        <Outlet />
      </main>
    </>
  )
}
