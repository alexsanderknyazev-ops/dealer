import { useEffect, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { Plus } from 'lucide-react'
import * as api from './supplierOrdersApi'
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
  linked: 'Связан с документом',
  fulfilled: 'Выполнен',
  cancelled: 'Отменён',
}

export function SupplierOrders() {
  const { logout } = useAuth()
  const navigate = useNavigate()
  const [list, setList] = useState<api.SupplierOrder[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [statusFilter, setStatusFilter] = useState('draft')
  const [page, setPage] = useState(0)
  const limit = 20

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    api
      .listSupplierOrders({ limit, offset: page * limit, status: statusFilter || undefined })
      .then((r) => {
        if (!cancelled) {
          setList(r.orders ?? [])
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
          setError(err instanceof Error ? err.message : 'Ошибка загрузки')
        }
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [page, statusFilter, logout, navigate])

  return (
    <div className="mx-auto w-full max-w-5xl space-y-6">
      <PageHeader
        title="Заказы поставщику"
        action={
          <Button asChild>
            <Link to="/supplier-orders/new">
              <Plus className="mr-1 h-4 w-4" /> Новый заказ
            </Link>
          </Button>
        }
      />
      {error && <ErrorAlert message={error} />}
      <Card>
        <CardContent className="flex flex-wrap items-center gap-3 pt-6">
          <NativeSelect value={statusFilter} onChange={(e) => { setPage(0); setStatusFilter(e.target.value) }}>
            <option value="">Все статусы</option>
            <option value="draft">Черновик</option>
            <option value="linked">Связан</option>
            <option value="fulfilled">Выполнен</option>
            <option value="cancelled">Отменён</option>
          </NativeSelect>
        </CardContent>
      </Card>
      {loading ? (
        <LoadingState />
      ) : list.length === 0 ? (
        <EmptyState>Заказов нет</EmptyState>
      ) : (
        <Card>
          <CardContent className="p-0">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Номер</TableHead>
                  <TableHead>Поставщик</TableHead>
                  <TableHead>Склад</TableHead>
                  <TableHead>Статус</TableHead>
                  <TableHead />
                </TableRow>
              </TableHeader>
              <TableBody>
                {list.map((o) => (
                  <TableRow key={o.id}>
                    <TableCell className="font-medium">{o.order_number}</TableCell>
                    <TableCell>{o.supplier_name || '—'}</TableCell>
                    <TableCell>{o.receipt_warehouse_name || '—'}</TableCell>
                    <TableCell>
                      <Badge variant="secondary">{STATUS_LABEL[o.status] || o.status}</Badge>
                    </TableCell>
                    <TableCell className="text-right">
                      <Link className="text-primary underline" to={`/supplier-orders/${o.id}`}>Открыть</Link>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      )}
      <Pagination page={page} limit={limit} total={total} onPageChange={setPage} />
    </div>
  )
}
