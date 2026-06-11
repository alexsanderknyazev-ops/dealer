import { useEffect, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { BarChart3, Car, Handshake, Package, Users } from 'lucide-react'
import { useAuth } from './auth'
import * as statsApi from '@/statsApi'
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
  const [employee, setEmployee] = useState<statsApi.EmployeeOverview | null>(null)
  const [client, setClient] = useState<statsApi.ClientOverview | null>(null)

  useEffect(() => {
    Promise.all([statsApi.getEmployeeOverview(), statsApi.getClientOverview()])
      .then(([emp, cli]) => {
        setEmployee(emp)
        setClient(cli)
      })
      .catch(() => {
        /* dashboard works without stats */
      })
  }, [])

  return (
    <div className="mx-auto max-w-3xl space-y-6">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight">Главная</h1>
        <p className="text-muted-foreground">Добро пожаловать в панель управления</p>
      </div>

      {(employee || client) && (
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0">
            <div>
              <CardTitle className="text-base">Сводка</CardTitle>
              <CardDescription>Актуальные показатели из сервисов статистики</CardDescription>
            </div>
            <Button variant="outline" size="sm" asChild>
              <Link to="/statistics">
                <BarChart3 className="mr-2 h-4 w-4" />
                Подробнее
              </Link>
            </Button>
          </CardHeader>
          <CardContent className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
            {employee && (
              <>
                <div>
                  <p className="text-xs text-muted-foreground">Клиенты</p>
                  <p className="text-xl font-semibold tabular-nums">{employee.customers_count}</p>
                </div>
                <div>
                  <p className="text-xs text-muted-foreground">Сделки</p>
                  <p className="text-xl font-semibold tabular-nums">{employee.deals_count}</p>
                </div>
              </>
            )}
            {client && (
              <>
                <div>
                  <p className="text-xs text-muted-foreground">B2C клиенты</p>
                  <p className="text-xl font-semibold tabular-nums">{client.clients_count}</p>
                </div>
                <div>
                  <p className="text-xs text-muted-foreground">Отзывы</p>
                  <p className="text-xl font-semibold tabular-nums">{client.reviews_count}</p>
                </div>
              </>
            )}
          </CardContent>
        </Card>
      )}

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
