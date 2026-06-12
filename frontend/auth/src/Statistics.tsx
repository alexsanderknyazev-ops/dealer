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
  X,
} from 'lucide-react'
import * as statsApi from '@/statsApi'
import * as reviewsApi from '@/reviewsApi'
import * as dealerPointsApi from '@/dealerPointsApi'
import { PageHeader } from '@/components/common/PageHeader'
import { LoadingState } from '@/components/common/LoadingState'
import { ErrorAlert } from '@/components/common/ErrorAlert'
import { Pagination } from '@/components/common/Pagination'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { NativeSelect } from '@/components/ui/native-select'
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

function formatDateTime(ts: number) {
  return ts ? new Date(ts * 1000).toLocaleString('ru-RU') : '—'
}

function ratingStars(rating: number) {
  const n = Math.max(0, Math.min(5, Math.round(rating)))
  return '★'.repeat(n) + '☆'.repeat(5 - n)
}

function truncate(text: string, max = 60) {
  const t = text.trim()
  if (t.length <= max) return t
  return `${t.slice(0, max)}…`
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
  const [statusFilter, setStatusFilter] = useState('')

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
          <CardDescription>Нажмите на строку, чтобы отфильтровать список ниже</CardDescription>
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
                  <TableRow
                    key={row.status}
                    className="cursor-pointer hover:bg-muted/50"
                    onClick={() => setStatusFilter((prev) => (prev === row.status ? '' : row.status))}
                  >
                    <TableCell>
                      <Badge variant={statusFilter === row.status ? 'default' : 'secondary'}>
                        {REVIEW_STATUS_LABEL[row.status] || row.status}
                      </Badge>
                    </TableCell>
                    <TableCell className="text-right tabular-nums">{row.count}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      <ClientReviewsList statusFilter={statusFilter} onStatusFilterChange={setStatusFilter} />
    </div>
  )
}

function ClientReviewsList({
  statusFilter,
  onStatusFilterChange,
}: {
  statusFilter: string
  onStatusFilterChange: (status: string) => void
}) {
  const [reviews, setReviews] = useState<reviewsApi.EmployeeReview[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(0)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [selected, setSelected] = useState<reviewsApi.EmployeeReview | null>(null)
  const [dealerPoints, setDealerPoints] = useState<dealerPointsApi.DealerPoint[]>([])
  const [dealerPointFilter, setDealerPointFilter] = useState('')
  const limit = 20

  const pointName = (id: string) => dealerPoints.find((p) => p.id === id)?.name || id || '—'

  const load = useCallback(() => {
    setLoading(true)
    setError(null)
    reviewsApi
      .listReviews({
        limit,
        offset: page * limit,
        status: statusFilter || undefined,
        dealer_point_id: dealerPointFilter || undefined,
      })
      .then((r) => {
        setReviews(r.reviews)
        setTotal(r.total)
      })
      .catch((e) => setError(e instanceof Error ? e.message : 'Ошибка загрузки отзывов'))
      .finally(() => setLoading(false))
  }, [page, statusFilter, dealerPointFilter])

  useEffect(() => {
    dealerPointsApi.listDealerPoints({ limit: 200 }).then((r) => setDealerPoints(r.dealer_points)).catch(() => {})
  }, [])

  useEffect(() => {
    load()
  }, [load])

  useEffect(() => {
    setPage(0)
  }, [statusFilter, dealerPointFilter])

  return (
    <div className="space-y-4">
      <Card>
        <CardHeader>
          <CardTitle className="text-base">Все отзывы</CardTitle>
          <CardDescription>Подробный список из read-модели employee-reviews</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex flex-wrap gap-3">
            <NativeSelect
              className="max-w-[220px]"
              value={statusFilter}
              onChange={(e) => onStatusFilterChange(e.target.value)}
            >
              <option value="">Все статусы</option>
              <option value="pending">На модерации</option>
              <option value="published">Опубликован</option>
              <option value="rejected">Отклонён</option>
            </NativeSelect>
            <NativeSelect
              className="max-w-[260px]"
              value={dealerPointFilter}
              onChange={(e) => setDealerPointFilter(e.target.value)}
            >
              <option value="">Все дилерские точки</option>
              {dealerPoints.map((p) => (
                <option key={p.id} value={p.id}>{p.name}</option>
              ))}
            </NativeSelect>
          </div>

          {error && <ErrorAlert message={error} onRetry={load} />}

          {loading ? (
            <LoadingState />
          ) : reviews.length === 0 ? (
            <p className="py-8 text-center text-sm text-muted-foreground">Отзывы не найдены</p>
          ) : (
            <>
              <div className="overflow-x-auto rounded-md border">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Дата</TableHead>
                      <TableHead>Клиент</TableHead>
                      <TableHead>Автомобиль</TableHead>
                      <TableHead>Оценка</TableHead>
                      <TableHead>Статус</TableHead>
                      <TableHead>Текст</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {reviews.map((r) => (
                      <TableRow
                        key={r.id}
                        className={cn('cursor-pointer hover:bg-muted/50', selected?.id === r.id && 'bg-muted/60')}
                        onClick={() => setSelected(r)}
                      >
                        <TableCell className="whitespace-nowrap text-sm">{formatDateTime(r.occurred_at || r.created_at)}</TableCell>
                        <TableCell>
                          <div className="font-medium">{r.client_full_name || '—'}</div>
                          {r.client_email && <div className="text-xs text-muted-foreground">{r.client_email}</div>}
                        </TableCell>
                        <TableCell className="text-sm">
                          {r.vehicle_make || r.vehicle_model
                            ? `${r.vehicle_make} ${r.vehicle_model}${r.vehicle_year ? ` (${r.vehicle_year})` : ''}`
                            : '—'}
                          {r.vehicle_vin && <div className="font-mono text-xs text-muted-foreground">{r.vehicle_vin}</div>}
                        </TableCell>
                        <TableCell className="whitespace-nowrap text-amber-600">{ratingStars(r.rating)}</TableCell>
                        <TableCell>
                          <Badge variant="secondary">{REVIEW_STATUS_LABEL[r.status] || r.status}</Badge>
                        </TableCell>
                        <TableCell className="max-w-[200px] text-sm text-muted-foreground">
                          {r.text ? truncate(r.text) : '—'}
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>
              <Pagination page={page} total={total} limit={limit} onPageChange={setPage} showTotal />
            </>
          )}
        </CardContent>
      </Card>

      {selected && (
        <Card>
          <CardHeader className="flex flex-row items-start justify-between gap-4">
            <div>
              <CardTitle className="text-base">Отзыв клиента</CardTitle>
              <CardDescription>{formatDateTime(selected.occurred_at || selected.created_at)}</CardDescription>
            </div>
            <Button variant="ghost" size="icon" onClick={() => setSelected(null)} aria-label="Закрыть">
              <X className="h-4 w-4" />
            </Button>
          </CardHeader>
          <CardContent className="grid gap-4 text-sm md:grid-cols-2">
            <div className="space-y-2">
              <div>
                <span className="text-muted-foreground">Клиент: </span>
                {selected.client_full_name || '—'}
              </div>
              <div>
                <span className="text-muted-foreground">Email: </span>
                {selected.client_email || '—'}
              </div>
              <div>
                <span className="text-muted-foreground">ID клиента: </span>
                <span className="font-mono text-xs">{selected.client_id}</span>
              </div>
              <div>
                <span className="text-muted-foreground">Дилерская точка: </span>
                {pointName(selected.dealer_point_id)}
              </div>
            </div>
            <div className="space-y-2">
              <div>
                <span className="text-muted-foreground">Автомобиль: </span>
                {selected.vehicle_make} {selected.vehicle_model}
                {selected.vehicle_year ? ` (${selected.vehicle_year})` : ''}
              </div>
              <div>
                <span className="text-muted-foreground">VIN: </span>
                <span className="font-mono text-xs">{selected.vehicle_vin || '—'}</span>
              </div>
              <div>
                <span className="text-muted-foreground">Оценка: </span>
                <span className="text-amber-600">{ratingStars(selected.rating)}</span>
                <span className="ml-1 text-muted-foreground">({selected.rating}/5)</span>
              </div>
              <div>
                <span className="text-muted-foreground">Статус: </span>
                <Badge variant="secondary">{REVIEW_STATUS_LABEL[selected.status] || selected.status}</Badge>
              </div>
            </div>
            <div className="md:col-span-2">
              <p className="mb-1 text-muted-foreground">Текст отзыва</p>
              <p className="whitespace-pre-wrap rounded-md border bg-muted/30 p-3">{selected.text || '—'}</p>
            </div>
            <div className="text-xs text-muted-foreground md:col-span-2">
              Review ID: <span className="font-mono">{selected.review_id}</span>
              {' · '}
              Создан: {formatDateTime(selected.created_at)}
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  )
}
