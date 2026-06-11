import { useCallback, useEffect, useState, type ComponentType } from 'react'
import {
  BarChart3,
  Car,
  Handshake,
  MapPin,
  MessageSquare,
  Package,
  RefreshCw,
  Star,
  Users,
} from 'lucide-react'
import * as statsApi from '@/statsApi'
import { PageHeader } from '@/components/common/PageHeader'
import { LoadingState } from '@/components/common/LoadingState'
import { ErrorAlert } from '@/components/common/ErrorAlert'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { cn } from '@/lib/utils'

const DEAL_STAGE_LABEL: Record<string, string> = {
  draft: 'Черновик',
  in_progress: 'В работе',
  paid: 'Оплачена',
  completed: 'Завершена',
  cancelled: 'Отменена',
}

const REVIEW_STATUS_LABEL: Record<string, string> = {
  pending: 'На модерации',
  published: 'Опубликован',
  rejected: 'Отклонён',
}

function formatMoney(value: number): string {
  return new Intl.NumberFormat('ru-RU', {
    style: 'currency',
    currency: 'RUB',
    maximumFractionDigits: 0,
  }).format(value)
}

function StatCard({
  label,
  value,
  hint,
  icon: Icon,
}: {
  label: string
  value: string | number
  hint?: string
  icon: ComponentType<{ className?: string }>
}) {
  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-sm font-medium text-muted-foreground">{label}</CardTitle>
        <Icon className="h-4 w-4 text-muted-foreground" />
      </CardHeader>
      <CardContent>
        <div className="text-2xl font-semibold tabular-nums">{value}</div>
        {hint && <p className="mt-1 text-xs text-muted-foreground">{hint}</p>}
      </CardContent>
    </Card>
  )
}

type Tab = 'employee' | 'client'

export function Statistics() {
  const [tab, setTab] = useState<Tab>('employee')
  const [employee, setEmployee] = useState<statsApi.EmployeeOverview | null>(null)
  const [client, setClient] = useState<statsApi.ClientOverview | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const load = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const [emp, cli] = await Promise.all([statsApi.getEmployeeOverview(), statsApi.getClientOverview()])
      setEmployee(emp)
      setClient(cli)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Ошибка загрузки статистики')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    load()
  }, [load])

  return (
    <div className="mx-auto w-full max-w-5xl space-y-6">
      <PageHeader
        title="Статистика"
        subtitle={<span className="text-muted-foreground">Сводные показатели из Kafka read-моделей</span>}
        action={
          <Button variant="outline" size="sm" onClick={load} disabled={loading}>
            <RefreshCw className={cn('mr-2 h-4 w-4', loading && 'animate-spin')} />
            Обновить
          </Button>
        }
      />

      <div className="flex gap-2">
        <Button variant={tab === 'employee' ? 'default' : 'outline'} size="sm" onClick={() => setTab('employee')}>
          Дилер
        </Button>
        <Button variant={tab === 'client' ? 'default' : 'outline'} size="sm" onClick={() => setTab('client')}>
          Клиентская зона
        </Button>
      </div>

      {error && <ErrorAlert message={error} onRetry={load} />}
      {loading && !employee && !client ? (
        <LoadingState />
      ) : tab === 'employee' && employee ? (
        <EmployeeStats overview={employee} />
      ) : tab === 'client' && client ? (
        <ClientStats overview={client} />
      ) : null}
    </div>
  )
}

function EmployeeStats({ overview }: { overview: statsApi.EmployeeOverview }) {
  return (
    <div className="space-y-6">
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        <StatCard label="Клиенты" value={overview.customers_count} icon={Users} />
        <StatCard label="Автомобили" value={overview.vehicles_count} icon={Car} />
        <StatCard label="Сделки" value={overview.deals_count} icon={Handshake} />
        <StatCard label="Запчасти" value={overview.parts_count} icon={Package} />
        <StatCard label="Дилерские точки" value={overview.dealer_points_count} icon={MapPin} />
        <StatCard
          label="Выручка"
          value={formatMoney(overview.total_revenue)}
          hint="По завершённым и оплаченным сделкам"
          icon={BarChart3}
        />
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Сделки по этапам</CardTitle>
          <CardDescription>Распределение по воронке продаж</CardDescription>
        </CardHeader>
        <CardContent className="p-0">
          {overview.deals_by_stage.length === 0 ? (
            <p className="px-6 py-8 text-center text-sm text-muted-foreground">Нет данных по этапам</p>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Этап</TableHead>
                  <TableHead className="text-right">Количество</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {overview.deals_by_stage.map((row) => (
                  <TableRow key={row.stage}>
                    <TableCell>
                      <Badge variant="secondary">{DEAL_STAGE_LABEL[row.stage] || row.stage}</Badge>
                    </TableCell>
                    <TableCell className="text-right tabular-nums">{row.count}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>
    </div>
  )
}

function ClientStats({ overview }: { overview: statsApi.ClientOverview }) {
  return (
    <div className="space-y-6">
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        <StatCard label="Клиенты (B2C)" value={overview.clients_count} icon={Users} />
        <StatCard label="Зарегистрированные" value={overview.registered_users_count} icon={Users} />
        <StatCard label="Привязанные авто" value={overview.client_vehicles_count} icon={Car} />
        <StatCard label="Отзывы" value={overview.reviews_count} icon={MessageSquare} />
        <StatCard
          label="Средняя оценка"
          value={overview.reviews_count > 0 ? overview.average_rating.toFixed(1) : '—'}
          hint={overview.reviews_count > 0 ? '★'.repeat(Math.min(5, Math.round(overview.average_rating))) : undefined}
          icon={Star}
        />
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Отзывы по статусам</CardTitle>
          <CardDescription>Модерация и публикация</CardDescription>
        </CardHeader>
        <CardContent className="p-0">
          {overview.reviews_by_status.length === 0 ? (
            <p className="px-6 py-8 text-center text-sm text-muted-foreground">Отзывов пока нет</p>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Статус</TableHead>
                  <TableHead className="text-right">Количество</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {overview.reviews_by_status.map((row) => (
                  <TableRow key={row.status}>
                    <TableCell>
                      <Badge variant="secondary">{REVIEW_STATUS_LABEL[row.status] || row.status}</Badge>
                    </TableCell>
                    <TableCell className="text-right tabular-nums">{row.count}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
