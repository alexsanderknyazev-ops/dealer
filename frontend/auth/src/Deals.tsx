import { useEffect, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { Plus } from 'lucide-react'
import type { Deal } from './dealsApi'
import * as api from './dealsApi'
import { useAuth } from './auth'
import { PageHeader } from '@/components/common/PageHeader'
import { ErrorAlert } from '@/components/common/ErrorAlert'
import { LoadingState } from '@/components/common/LoadingState'
import { EmptyState } from '@/components/common/EmptyState'
import { Pagination } from '@/components/common/Pagination'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { NativeSelect } from '@/components/ui/native-select'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'

const STAGE_LABEL: Record<string, string> = {
  draft: 'Черновик',
  in_progress: 'В работе',
  paid: 'Оплачено',
  completed: 'Завершено',
  cancelled: 'Отменено',
}

export function Deals() {
  const { logout } = useAuth()
  const navigate = useNavigate()
  const [list, setList] = useState<Deal[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [stageFilter, setStageFilter] = useState('')
  const [page, setPage] = useState(0)
  const [retry, setRetry] = useState(0)
  const limit = 20

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    setError(null)
    api
      .listDeals({ limit, offset: page * limit, stage: stageFilter || undefined })
      .then((r) => {
        if (!cancelled) {
          setList(r.deals)
          setTotal(r.total)
        }
      })
      .catch(async (err) => {
        if (!cancelled) {
          if (err instanceof api.ApiError && (err.status === 401 || err.status === 403)) {
            await logout()
            navigate('/login', { replace: true })
            return
          }
          setList([])
          setError(err instanceof Error ? err.message : 'Ошибка загрузки')
        }
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [page, stageFilter, retry, logout, navigate])

  return (
    <div className="mx-auto w-full max-w-5xl">
      <PageHeader
        title="Сделки"
        action={
          <Button asChild>
            <Link to="/deals/new">
              <Plus className="mr-2 h-4 w-4" />
              Новая сделка
            </Link>
          </Button>
        }
      />

      <div className="mb-4 max-w-xs">
        <NativeSelect
          value={stageFilter}
          onChange={(e) => {
            setStageFilter(e.target.value)
            setPage(0)
          }}
        >
          <option value="">Все этапы</option>
          <option value="draft">Черновик</option>
          <option value="in_progress">В работе</option>
          <option value="paid">Оплачено</option>
          <option value="completed">Завершено</option>
          <option value="cancelled">Отменено</option>
        </NativeSelect>
      </div>

      {error && <ErrorAlert message={error} onRetry={() => setRetry((r) => r + 1)} />}

      {loading ? (
        <LoadingState />
      ) : list.length === 0 && !error ? (
        <EmptyState>Нет сделок. Нажмите «Новая сделка» для создания.</EmptyState>
      ) : (
        <>
          <Card>
            <CardContent className="p-0">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Клиент ID</TableHead>
                    <TableHead>Авто ID</TableHead>
                    <TableHead>Сумма</TableHead>
                    <TableHead>Этап</TableHead>
                    <TableHead />
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {list.map((d) => (
                    <TableRow key={d.id}>
                      <TableCell className="font-mono text-xs">{d.customer_id.slice(0, 8)}…</TableCell>
                      <TableCell className="font-mono text-xs">{d.vehicle_id.slice(0, 8)}…</TableCell>
                      <TableCell>{d.amount ? Number(d.amount).toLocaleString('ru') : '—'}</TableCell>
                      <TableCell>
                        <Badge variant="secondary">{STAGE_LABEL[d.stage] || d.stage}</Badge>
                      </TableCell>
                      <TableCell>
                        <Button variant="link" className="h-auto p-0" asChild>
                          <Link to={`/deals/${d.id}`}>Открыть</Link>
                        </Button>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </CardContent>
          </Card>
          <Pagination page={page} total={total} limit={limit} onPageChange={setPage} />
        </>
      )}
    </div>
  )
}
