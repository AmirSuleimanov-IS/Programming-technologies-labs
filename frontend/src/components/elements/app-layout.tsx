import { NavLink, Outlet } from 'react-router'

const APP_TITLE = import.meta.env.VITE_APP_TITLE ?? 'SportReg'

export function AppLayout() {
  return (
    <>
      <header className="header">
        <div className="container header__inner">
          <NavLink to="/" className="logo">
            <img src="/favicon.svg" alt="" width={28} height={28} />
            {APP_TITLE}
          </NavLink>
          <nav className="nav">
            <NavLink to="/competitions">Соревнования</NavLink>
          </nav>
        </div>
      </header>
      <main className="container main">
        <Outlet />
      </main>
    </>
  )
}
