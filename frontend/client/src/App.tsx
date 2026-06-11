import type { ReactNode } from 'react'
import { Navigate, Route, Routes } from 'react-router-dom'
import { useAuth } from '@/auth'
import { AppShell } from '@/components/layout/AppShell'
import { Login } from '@/pages/Login'
import { Register } from '@/pages/Register'
import { Dashboard } from '@/pages/Dashboard'
import { Vehicles } from '@/pages/Vehicles'
import { Reviews } from '@/pages/Reviews'
import { ReviewNew } from '@/pages/ReviewNew'

function RequireAuth({ children }: { children: ReactNode }) {
  const { user, loading } = useAuth()
  if (loading) {
    return <div className="flex min-h-[50vh] items-center justify-center text-muted-foreground">Загрузка…</div>
  }
  if (!user) return <Navigate to="/login" replace />
  return <>{children}</>
}

function GuestOnly({ children }: { children: ReactNode }) {
  const { user, loading } = useAuth()
  if (loading) {
    return <div className="flex min-h-[50vh] items-center justify-center text-muted-foreground">Загрузка…</div>
  }
  if (user) return <Navigate to="/" replace />
  return <>{children}</>
}

export default function App() {
  return (
    <Routes>
      <Route path="/" element={<AppShell />}>
        <Route index element={<RequireAuth><Dashboard /></RequireAuth>} />
        <Route path="vehicles" element={<RequireAuth><Vehicles /></RequireAuth>} />
        <Route path="reviews" element={<RequireAuth><Reviews /></RequireAuth>} />
        <Route path="reviews/new" element={<RequireAuth><ReviewNew /></RequireAuth>} />
        <Route path="login" element={<GuestOnly><Login /></GuestOnly>} />
        <Route path="register" element={<GuestOnly><Register /></GuestOnly>} />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Route>
    </Routes>
  )
}
