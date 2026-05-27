import { Link, Outlet } from 'react-router-dom'
import { useAuth } from './auth'
import './Layout.css'

export function Layout() {
  const { user } = useAuth()

  return (
    <div className="layout">
      <header className="header">
        <Link to="/" className="logo">Dealer</Link>
        <nav className="header-nav">
          {user && (
            <>
              <Link to="/customers" className="header-link">Клиенты</Link>
              <Link to="/vehicles" className="header-link">Автомобили</Link>
              <Link to="/deals" className="header-link">Сделки</Link>
              <Link to="/parts" className="header-link">Запчасти</Link>
              <Link to="/brands" className="header-link">Бренды</Link>
              <Link to="/dealer-points" className="header-link">Дилерские точки</Link>
              <Link to="/legal-entities" className="header-link">Юр. лица</Link>
              <Link to="/warehouses" className="header-link">Склады</Link>
              <span className="user-email">{user.email}</span>
            </>
          )}
        </nav>
      </header>
      <main className="main">
        <Outlet />
      </main>
    </div>
  )
}
