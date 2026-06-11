import { useNavigate } from 'react-router-dom'
import { Car, Handshake, Package, Users } from 'lucide-react'
import { useAuth } from './auth'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'

const quickLinks = [
  { to: '/customers', label: 'Клиенты', icon: Users },
  { to: '/vehicles', label: 'Автомобили', icon: Car },
  { to: '/deals', label: 'Сделки', icon: Handshake },
  { to: '/parts', label: 'Запчасти', icon: Package },
]

export function Dashboard() {
  const { user } = useAuth()
  const navigate = useNavigate()

  return (
    <div className="mx-auto max-w-2xl space-y-6">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight">Главная</h1>
        <p className="text-muted-foreground">Добро пожаловать в панель управления</p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Вы вошли как</CardTitle>
          <CardDescription className="text-base text-foreground">{user?.email}</CardDescription>
        </CardHeader>
        <CardContent className="grid gap-2 sm:grid-cols-2">
          {quickLinks.map(({ to, label, icon: Icon }) => (
            <Button key={to} variant="outline" className="justify-start" onClick={() => navigate(to)}>
              <Icon className="mr-2 h-4 w-4" />
              {label}
            </Button>
          ))}
        </CardContent>
      </Card>
    </div>
  )
}
