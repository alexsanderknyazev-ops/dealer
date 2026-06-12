import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import * as dealerPointsApi from './dealerPointsApi'
import * as reviewsApi from './reviewsApi'
import { PageHeader } from '@/components/common/PageHeader'
import { LoadingState } from '@/components/common/LoadingState'
import { ErrorAlert } from '@/components/common/ErrorAlert'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { REVIEW_STATUS_LABEL, formatReviewDateTime, ratingStars } from './reviewsShared'

export function ReviewView() {
  const { id } = useParams()
  const [review, setReview] = useState<reviewsApi.EmployeeReview | null>(null)
  const [dealerPointName, setDealerPointName] = useState('—')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!id) return
    let cancelled = false
    setLoading(true)
    setError(null)
    reviewsApi
      .getReview(id)
      .then((r) => {
        if (!cancelled) setReview(r)
        return dealerPointsApi.listDealerPoints({ limit: 200 }).then((pts) => {
          const point = pts.dealer_points.find((p) => p.id === r.dealer_point_id)
          if (!cancelled) setDealerPointName(point?.name || r.dealer_point_id || '—')
        })
      })
      .catch((e) => {
        if (!cancelled) setError(e instanceof Error ? e.message : 'Ошибка загрузки')
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [id])

  if (loading) return <LoadingState />
  if (error) return <ErrorAlert message={error} />
  if (!review) return <ErrorAlert message="Отзыв не найден" />

  const backTo = review.dealer_point_id
    ? `/reviews?dealer_point_id=${review.dealer_point_id}`
    : '/reviews'

  return (
    <div className="mx-auto w-full max-w-2xl space-y-4">
      <PageHeader
        title="Отзыв клиента"
        subtitle={formatReviewDateTime(review.occurred_at || review.created_at)}
        action={
          <Button variant="outline" size="sm" asChild>
            <Link to={backTo}>К списку отзывов</Link>
          </Button>
        }
      />

      <Card>
        <CardHeader>
          <div className="flex flex-wrap items-center gap-2">
            <CardTitle className="text-lg">{review.client_full_name || 'Клиент'}</CardTitle>
            <Badge variant="secondary">{REVIEW_STATUS_LABEL[review.status] || review.status}</Badge>
          </div>
          <CardDescription>{review.client_email || 'Email не указан'}</CardDescription>
        </CardHeader>
        <CardContent className="grid gap-4 text-sm md:grid-cols-2">
          <div className="space-y-2">
            <div>
              <span className="text-muted-foreground">Дилерская точка: </span>
              {dealerPointName}
            </div>
            <div>
              <span className="text-muted-foreground">ID клиента: </span>
              <span className="font-mono text-xs">{review.client_id}</span>
            </div>
          </div>
          <div className="space-y-2">
            <div>
              <span className="text-muted-foreground">Автомобиль: </span>
              {review.vehicle_make} {review.vehicle_model}
              {review.vehicle_year ? ` (${review.vehicle_year})` : ''}
            </div>
            <div>
              <span className="text-muted-foreground">VIN: </span>
              <span className="font-mono text-xs">{review.vehicle_vin || '—'}</span>
            </div>
            <div>
              <span className="text-muted-foreground">Оценка: </span>
              <span className="text-amber-600">{ratingStars(review.rating)}</span>
              <span className="ml-1 text-muted-foreground">({review.rating}/5)</span>
            </div>
          </div>
          <div className="md:col-span-2">
            <p className="mb-1 text-muted-foreground">Текст отзыва</p>
            <p className="whitespace-pre-wrap rounded-md border bg-muted/30 p-3">{review.text || '—'}</p>
          </div>
          <div className="text-xs text-muted-foreground md:col-span-2">
            Review ID: <span className="font-mono">{review.review_id}</span>
            {' · '}
            Создан: {formatReviewDateTime(review.created_at)}
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
