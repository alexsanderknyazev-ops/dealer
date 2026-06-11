import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { Plus } from 'lucide-react'
import { useAuth } from '@/auth'
import * as api from '@/api'
import { PageHeader } from '@/components/common/PageHeader'
import { LoadingState } from '@/components/common/LoadingState'
import { ErrorAlert } from '@/components/common/ErrorAlert'
import { EmptyState } from '@/components/common/EmptyState'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'

const STATUS_LABEL: Record<string, string> = {
  pending: 'На модерации',
  published: 'Опубликован',
  rejected: 'Отклонён',
}

export function Reviews() {
  const { getAccessToken } = useAuth()
  const [list, setList] = useState<api.Review[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    getAccessToken()
      .then((token) => {
        if (!token) throw new Error('Сессия истекла')
        return api.listReviews(token)
      })
      .then((r) => {
        if (!cancelled) setList(r.reviews ?? [])
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
  }, [getAccessToken])

  return (
    <div className="mx-auto w-full max-w-4xl space-y-4">
      <PageHeader
        title="Мои отзывы"
        action={
          <Button asChild>
            <Link to="/reviews/new">
              <Plus className="mr-2 h-4 w-4" />
              Новый отзыв
            </Link>
          </Button>
        }
      />

      {error && <ErrorAlert message={error} />}
      {loading ? (
        <LoadingState />
      ) : list.length === 0 && !error ? (
        <EmptyState>Отзывов пока нет. Оставьте первый отзыв о визите в дилерский центр.</EmptyState>
      ) : (
        <Card>
          <CardContent className="p-0">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Оценка</TableHead>
                  <TableHead>Текст</TableHead>
                  <TableHead>Статус</TableHead>
                  <TableHead>Дата</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {list.map((r) => (
                  <TableRow key={r.id}>
                    <TableCell>{'★'.repeat(Math.min(5, Math.max(0, r.rating)))}</TableCell>
                    <TableCell className="max-w-md truncate">{r.text || '—'}</TableCell>
                    <TableCell>
                      <Badge variant="secondary">{STATUS_LABEL[r.status] || r.status}</Badge>
                    </TableCell>
                    <TableCell>{r.created_at ? new Date(r.created_at * 1000).toLocaleDateString('ru') : '—'}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      )}
    </div>
  )
}
