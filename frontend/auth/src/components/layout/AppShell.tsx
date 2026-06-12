import { Link, NavLink, Outlet, useLocation, useNavigate } from 'react-router-dom'
import {
  BarChart3,
  Building2,
  Car,
  Handshake,
  ArrowLeftRight,
  ClipboardList,
  Clock,
  Wrench,
  Home,
  LogOut,
  MapPin,
  Package,
  Scale,
  Tag,
  Users,
  Warehouse,
} from 'lucide-react'
import { useAuth } from '@/auth'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'
import { ThemeToggle } from '@/components/layout/ThemeToggle'

const navItems = [
  { to: '/', label: 'Главная', icon: Home, end: true },
  { to: '/customers', label: 'Клиенты', icon: Users },
  { to: '/vehicles', label: 'Автомобили', icon: Car },
  { to: '/deals', label: 'Сделки', icon: Handshake },
  { to: '/work-orders', label: 'Заказ-наряды', icon: ClipboardList },
  { to: '/movement-documents', label: 'Перемещение товаров', icon: ArrowLeftRight },
  { to: '/parts', label: 'Запчасти', icon: Package },
  { to: '/works', label: 'Работы', icon: Wrench },
  { to: '/brands', label: 'Бренды', icon: Tag },
  { to: '/brand-labor-rates', label: 'Нормо-часы', icon: Clock },
  { to: '/dealer-points', label: 'Дилерские точки', icon: MapPin },
  { to: '/legal-entities', label: 'Юр. лица', icon: Scale },
  { to: '/warehouses', label: 'Склады', icon: Warehouse },
  { to: '/statistics', label: 'Статистика', icon: BarChart3 },
]

function GuestShell() {
  return (
    <div className="flex min-h-screen flex-col">
      <header className="flex items-center justify-between border-b px-4 py-3">
        <Link to="/" className="flex items-center gap-2 text-lg font-semibold tracking-tight">
          <Building2 className="h-5 w-5 text-primary" />
          Dealer
        </Link>
        <ThemeToggle />
      </header>
      <main className="flex flex-1 items-center justify-center p-4">
        <Outlet />
      </main>
    </div>
  )
}

export function AppShell() {
  const { user, logout } = useAuth()
  const navigate = useNavigate()
  const location = useLocation()
  const isAuthPage = location.pathname === '/login' || location.pathname === '/register'

  async function handleLogout() {
    await logout()
    navigate('/login', { replace: true })
  }

  if (isAuthPage || !user) {
    return <GuestShell />
  }

  return (
    <div className="flex min-h-screen">
      <aside className="hidden w-56 shrink-0 flex-col border-r border-sidebar-border bg-sidebar md:flex">
        <div className="flex h-14 items-center gap-2 border-b border-sidebar-border px-4">
          <Building2 className="h-5 w-5 text-primary" />
          <span className="font-semibold tracking-tight text-sidebar-foreground">Dealer</span>
        </div>
        <nav className="flex flex-1 flex-col gap-1 p-3">
          {navItems.map(({ to, label, icon: Icon, end }) => (
            <NavLink
              key={to}
              to={to}
              end={end}
              className={({ isActive }) =>
                cn(
                  'flex items-center gap-2 rounded-md px-3 py-2 text-sm transition-colors',
                  isActive
                    ? 'bg-sidebar-accent font-medium text-sidebar-accent-foreground'
                    : 'text-sidebar-foreground hover:bg-sidebar-accent/60 hover:text-sidebar-accent-foreground',
                )
              }
            >
              <Icon className="h-4 w-4 shrink-0" />
              {label}
            </NavLink>
          ))}
        </nav>
      </aside>

      <div className="flex min-w-0 flex-1 flex-col">
        <header className="flex h-14 items-center justify-between gap-3 border-b px-4">
          <nav className="flex items-center gap-1 overflow-x-auto md:hidden">
            {navItems.slice(0, 4).map(({ to, label, end }) => (
              <NavLink
                key={to}
                to={to}
                end={end}
                className={({ isActive }) =>
                  cn(
                    'whitespace-nowrap rounded-md px-2 py-1 text-xs',
                    isActive ? 'bg-accent font-medium' : 'text-muted-foreground',
                  )
                }
              >
                {label}
              </NavLink>
            ))}
          </nav>
          <div className="hidden md:block" />
          <div className="flex items-center gap-2">
            <span className="hidden max-w-[200px] truncate text-sm text-muted-foreground sm:inline">
              {user.email}
            </span>
            <ThemeToggle />
            <Separator orientation="vertical" className="h-6" />
            <Button variant="ghost" size="sm" onClick={handleLogout}>
              <LogOut className="mr-1 h-4 w-4" />
              Выйти
            </Button>
          </div>
        </header>
        <main className="flex-1 overflow-auto p-4 md:p-6">
          <Outlet />
        </main>
      </div>
    </div>
  )
}
