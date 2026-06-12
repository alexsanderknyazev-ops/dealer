import { useCallback, useEffect, useState } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { MapPin, Star } from 'lucide-react'
import type { DealerPoint } from './dealerPointsApi'
import * as dealerPointsApi from './dealerPointsApi'
import * as reviewsApi from './reviewsApi'
import { PageHeader } from '@/components/common/PageHeader'
import { ErrorAlert } from '@/components/common/ErrorAlert'
import { LoadingState } from '@/components/common/LoadingState'
import { EmptyState } from '@/components/common/EmptyState'
import { Pagination } from '@/components/common/Pagination'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { NativeSelect } from '@/components/ui/native-select'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { cn } from '@/lib/utils'
import {
  REVIEW_STATUS_LABEL,
  formatReviewDateTime,
  ratingStars,
  truncateReviewText,
} from './reviewsShared'

const limit = 20

export function Reviews() {
  const [searchParams, setSearchParams] = useSearchParams()
  const dealerPointFilter = searchParams.get('dealer_point_id') ?? ''
  const [statusFilter, setStatusFilter] = useState('')
  const [page, setPage] = useState(0)
  const [reviews, setReviews] = useState<reviewsApi.EmployeeReview[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [dealerPoints, setDealerPoints] = useState<DealerPoint[]>([])
  const [statsByPoint, setStatsByPoint] = useState<Record<string, reviewsApi.ReviewStats>>({})
  const [pointsLoading, setPointsLoading] = useState(true)

  const pointName = (id: string) => dealerPoints.find((p) => p.id === id)?.name || '—'

  useEffect(() => {
    let cancelled = false
    setPointsLoading(true)
    dealerPointsApi
      .listDealerPoints({ limit: 200 })
      .then(async (r) => {
        if (cancelled) return
        setDealerPoints(r.dealer_points)
        const entries = await Promise.all(
          r.dealer_points.map(async (p) => {
            try {
              const stats = await reviewsApi.getReviewStats({ dealer_point_id: p.id })
              return [p.id, stats] as const
            } catch {
              return [p.id, { total_count: 0, average_rating: 0, by_status: [] }] as const
            }
          }),
        )
        if (!cancelled) setStatsByPoint(Object.fromEntries(entries))
      })
      .catch(() => {})
      .finally(() => {
        if (!cancelled) setPointsLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [])

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
    load()
  }, [load])

  useEffect(() => {
    setPage(0)
  }, [statusFilter, dealerPointFilter])

  function setDealerPointFilter(id: string) {
    const next = new URLSearchParams(searchParams)
    if (id) next.set('dealer_point_id', id)
    else next.delete('dealer_point_id')
    setSearchParams(next, { replace: true })
  }

  const totalAll = Object.values(statsByPoint).reduce((sum, s) => sum + s.total_count, 0)

  return (
    <div className="mx-auto w-full max-w-5xl space-y-6">
      <PageHeader title="Отзывы" subtitle="Отзывы клиентов по дилерским точкам" />

      <div className="space-y-3">
        <h2 className="text-sm font-medium text-muted-foreground">По дилерским точкам</h2>
        {pointsLoading ? (
          <LoadingState />
        ) : (
          <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
            <button
              type="button"
              onClick={() => setDealerPointFilter('')}
              className={cn(
                'rounded-lg border p-4 text-left transition-colors hover:bg-muted/50',
                !dealerPointFilter && 'border-primary bg-primary/5',
              )}
            >
              <div className="flex items-center gap-2 font-medium">
                <MapPin className="h-4 w-4 text-muted-foreground" />
                Все точки
              </div>
              <p className="mt-2 text-2xl font-semibold tabular-nums">{totalAll}</p>
              <p className="text-xs text-muted-foreground">отзывов всего</p>
            </button>
            {dealerPoints.map((p) => {
              const stats = statsByPoint[p.id]
              const active = dealerPointFilter === p.id
              return (
                <button
                  key={p.id}
                  type="button"
                  onClick={() => setDealerPointFilter(p.id)}
                  className={cn(
                    'rounded-lg border p-4 text-left transition-colors hover:bg-muted/50',
                    active && 'border-primary bg-primary/5',
                  )}
                >
                  <div className="flex items-center gap-2 font-medium">
                    <MapPin className="h-4 w-4 shrink-0 text-muted-foreground" />
                    <span className="truncate">{p.name}</span>
                  </div>
                  <p className="mt-2 text-2xl font-semibold tabular-nums">{stats?.total_count ?? 0}</p>
                  <p className="text-xs text-muted-foreground">
                    {stats && stats.average_rating > 0
                      ? `средняя оценка ${stats.average_rating.toFixed(1)}`
                      : 'нет опубликованных оценок'}
                  </p>
                </button>
              )
            })}
          </div>
        )}
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">
            {dealerPointFilter ? `Отзывы: ${pointName(dealerPointFilter)}` : 'Все отзывы'}
          </CardTitle>
          <CardDescription>Нажмите на строку, чтобы открыть отзыв</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex flex-wrap gap-3">
            <NativeSelect
              className="max-w-[220px]"
              value={statusFilter}
              onChange={(e) => setStatusFilter(e.target.value)}
            >
              <option value="">Все статусы</option>
              <option value="pending">На модерации</option>
              <option value="published">Опубликован</option>
              <option value="rejected">Отклонён</option>
            </NativeSelect>
            {dealerPointFilter && (
              <button
                type="button"
                className="text-sm text-muted-foreground underline-offset-4 hover:underline"
                onClick={() => setDealerPointFilter('')}
              >
                Сбросить фильтр точки
              </button>
            )}
          </div>

          {error && <ErrorAlert message={error} onRetry={load} />}

          {loading ? (
            <LoadingState />
          ) : reviews.length === 0 && !error ? (
            <EmptyState>
              {dealerPointFilter || statusFilter
                ? 'Отзывы не найдены по выбранным фильтрам.'
                : 'Пока нет отзывов. Они появятся после публикации клиентами.'}
            </EmptyState>
          ) : (
            <>
              <div className="overflow-x-auto rounded-md border">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Дата</TableHead>
                      {!dealerPointFilter && <TableHead>Дилерская точка</TableHead>}
                      <TableHead>Клиент</TableHead>
                      <TableHead>Автомобиль</TableHead>
                      <TableHead>Оценка</TableHead>
                      <TableHead>Статус</TableHead>
                      <TableHead>Текст</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {reviews.map((r) => (
                      <TableRow key={r.id} className="hover:bg-muted/50">
                        <TableCell className="whitespace-nowrap text-sm">
                          <Link to={`/reviews/${r.id}`} className="block">
                            {formatReviewDateTime(r.occurred_at || r.created_at)}
                          </Link>
                        </TableCell>
                        {!dealerPointFilter && (
                          <TableCell className="text-sm">
                            <Link to={`/reviews/${r.id}`} className="block">
                              {pointName(r.dealer_point_id)}
                            </Link>
                          </TableCell>
                        )}
                        <TableCell>
                          <Link to={`/reviews/${r.id}`} className="block">
                            <div className="font-medium">{r.client_full_name || '—'}</div>
                            {r.client_email && (
                              <div className="text-xs text-muted-foreground">{r.client_email}</div>
                            )}
                          </Link>
                        </TableCell>
                        <TableCell className="text-sm">
                          <Link to={`/reviews/${r.id}`} className="block">
                            {r.vehicle_make || r.vehicle_model
                              ? `${r.vehicle_make} ${r.vehicle_model}${r.vehicle_year ? ` (${r.vehicle_year})` : ''}`
                              : '—'}
                            {r.vehicle_vin && (
                              <div className="font-mono text-xs text-muted-foreground">{r.vehicle_vin}</div>
                            )}
                          </Link>
                        </TableCell>
                        <TableCell className="whitespace-nowrap text-amber-600">
                          <Link to={`/reviews/${r.id}`} className="block">
                            {ratingStars(r.rating)}
                          </Link>
                        </TableCell>
                        <TableCell>
                          <Link to={`/reviews/${r.id}`} className="block">
                            <Badge variant="secondary">{REVIEW_STATUS_LABEL[r.status] || r.status}</Badge>
                          </Link>
                        </TableCell>
                        <TableCell className="max-w-[200px] text-sm text-muted-foreground">
                          <Link to={`/reviews/${r.id}`} className="block">
                            {r.text ? truncateReviewText(r.text) : '—'}
                          </Link>
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

      {!pointsLoading && totalAll === 0 && !loading && reviews.length === 0 && (
        <p className="flex items-center gap-2 text-sm text-muted-foreground">
          <Star className="h-4 w-4" />
          Клиенты оставляют отзывы в личном кабинете после обслуживания или покупки.
        </p>
      )}
    </div>
  )
}
