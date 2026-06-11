import { useEffect, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { Plus } from 'lucide-react'
import * as api from './workOrdersApi'
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

const STATUS_LABEL: Record<string, string> = {
  draft: 'Черновик',
  in_progress: 'В работе',
  completed: 'Выполнен',
  closed: 'Закрыт',
  paid: 'Оплачен',
}

const REPAIR_LABEL: Record<string, string> = {
  warranty_manufacturer: 'Гарантия производителя',
  pre_sale: 'Предпродажная подготовка',
  commercial: 'Коммерческий ремонт',
  maintenance: 'Техобслуживание',
}

export function WorkOrders() {
  const { logout } = useAuth()
  const navigate = useNavigate()
  const [list, setList] = useState<api.WorkOrder[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [statusFilter, setStatusFilter] = useState('')
  const [page, setPage] = useState(0)
  const [retry, setRetry] = useState(0)
  const limit = 20

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    setError(null)
    api
      .listWorkOrders({ limit, offset: page * limit, status: statusFilter || undefined })
      .then((r) => {
        if (!cancelled) {
          setList(r.work_orders)
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
  }, [page, statusFilter, retry, logout, navigate])

  return (
    <div className="mx-auto w-full max-w-6xl">
      <PageHeader
        title="Заказ-наряды"
        action={
          <Button asChild>
            <Link to="/work-orders/new">
              <Plus className="mr-2 h-4 w-4" />
              Новый заказ-наряд
            </Link>
          </Button>
        }
      />

      <div className="mb-4 max-w-xs">
        <NativeSelect
          value={statusFilter}
          onChange={(e) => {
            setStatusFilter(e.target.value)
            setPage(0)
          }}
        >
          <option value="">Все статусы</option>
          <option value="draft">Черновик</option>
          <option value="in_progress">В работе</option>
          <option value="completed">Выполнен</option>
          <option value="closed">Закрыт</option>
          <option value="paid">Оплачен</option>
        </NativeSelect>
      </div>

      {error && <ErrorAlert message={error} onRetry={() => setRetry((r) => r + 1)} />}

      {loading ? (
        <LoadingState />
      ) : list.length === 0 && !error ? (
        <EmptyState>Нет заказ-нарядов. Создайте первый.</EmptyState>
      ) : (
        <>
          <Card>
            <CardContent className="p-0">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Номер</TableHead>
                    <TableHead>Вид ремонта</TableHead>
                    <TableHead>Статус</TableHead>
                    <TableHead>Сумма</TableHead>
                    <TableHead>Открыт</TableHead>
                    <TableHead />
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {list.map((wo) => (
                    <TableRow key={wo.id}>
                      <TableCell className="font-medium">{wo.order_number}</TableCell>
                      <TableCell>{REPAIR_LABEL[wo.repair_type] || wo.repair_type}</TableCell>
                      <TableCell>
                        <Badge variant="secondary">{STATUS_LABEL[wo.status] || wo.status}</Badge>
                      </TableCell>
                      <TableCell>{wo.total_cost || '0'}</TableCell>
                      <TableCell>
                        {wo.opened_at ? new Date(wo.opened_at * 1000).toLocaleString('ru-RU') : '—'}
                      </TableCell>
                      <TableCell className="text-right">
                        <Button variant="ghost" size="sm" asChild>
                          <Link to={`/work-orders/${wo.id}`}>Открыть</Link>
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
