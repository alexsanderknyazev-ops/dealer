import { useEffect, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { Check, Plus } from 'lucide-react'
import * as api from './movementDocumentsApi'
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
  closed: 'Закрыт',
  confirmed: 'Закрыт',
  cancelled: 'Отменён',
}

const MOVEMENT_TYPE_LABEL: Record<string, string> = {
  work_order_issue: 'В работу',
  transfer: 'Между складами',
  to_production: 'В производство',
  from_production: 'Извлечение (возврат)',
}

export function MovementDocuments() {
  const { logout, user } = useAuth()
  const navigate = useNavigate()
  const [list, setList] = useState<api.MovementDocument[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [statusFilter, setStatusFilter] = useState('draft')
  const [page, setPage] = useState(0)
  const [retry, setRetry] = useState(0)
  const [confirmingId, setConfirmingId] = useState<string | null>(null)
  const limit = 20

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    setError(null)
    api
      .listMovementDocuments({
        limit,
        offset: page * limit,
        status: statusFilter || undefined,
      })
      .then((r) => {
        if (!cancelled) {
          setList(r.documents ?? [])
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

  async function handleStart(doc: api.MovementDocument) {
    setConfirmingId(doc.id)
    setError(null)
    try {
      await api.startMovementDocument(doc.id)
      setRetry((r) => r + 1)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Не удалось перевести в работу')
    } finally {
      setConfirmingId(null)
    }
  }

  async function handleClose(doc: api.MovementDocument) {
    if (
      !window.confirm(
        `Закрыть документ ${doc.document_number}? Запчасти будут списаны со склада.`,
      )
    ) {
      return
    }
    setConfirmingId(doc.id)
    setError(null)
    try {
      await api.closeMovementDocument(doc.id, user?.userId)
      setRetry((r) => r + 1)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Не удалось закрыть документ')
    } finally {
      setConfirmingId(null)
    }
  }

  return (
    <div className="mx-auto w-full max-w-6xl">
      <PageHeader
        title="Перемещение товаров"
        action={
          <Button asChild>
            <Link to="/movement-documents/new">
              <Plus className="mr-2 h-4 w-4" />
              Новый документ
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
          <option value="draft">Черновики</option>
          <option value="in_progress">В работе</option>
          <option value="closed">Закрытые</option>
          <option value="cancelled">Отменённые</option>
        </NativeSelect>
      </div>

      {error && <ErrorAlert message={error} onRetry={() => setRetry((r) => r + 1)} />}

      {loading ? (
        <LoadingState />
      ) : list.length === 0 && !error ? (
        <EmptyState>
          {statusFilter === 'draft'
            ? 'Нет черновиков.'
            : statusFilter === 'in_progress'
              ? 'Нет документов в работе.'
              : 'Нет документов перемещения.'}
        </EmptyState>
      ) : (
        <>
          <Card>
            <CardContent className="p-0">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Номер</TableHead>
                    <TableHead>Тип</TableHead>
                    <TableHead>Статус</TableHead>
                    <TableHead>Основание</TableHead>
                    <TableHead>Строк</TableHead>
                    <TableHead>Создан</TableHead>
                    <TableHead />
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {list.map((doc) => (
                    <TableRow key={doc.id}>
                      <TableCell className="font-medium">{doc.document_number}</TableCell>
                      <TableCell>
                        {MOVEMENT_TYPE_LABEL[doc.movement_type] || doc.movement_type}
                      </TableCell>
                      <TableCell>
                        <Badge variant={doc.status === 'draft' ? 'default' : 'secondary'}>
                          {STATUS_LABEL[doc.status] || doc.status}
                        </Badge>
                      </TableCell>
                      <TableCell>
                        {doc.reference_type === 'work_order' && doc.reference_id ? (
                          <Link className="text-primary underline" to={`/work-orders/${doc.reference_id}`}>
                            {doc.reference_label || 'Заказ-наряд'}
                          </Link>
                        ) : (
                          doc.reference_type || '—'
                        )}
                      </TableCell>
                      <TableCell>{doc.lines?.length ?? 0}</TableCell>
                      <TableCell>
                        {doc.created_at > 0
                          ? new Date(doc.created_at * 1000).toLocaleString('ru-RU')
                          : '—'}
                      </TableCell>
                      <TableCell className="text-right">
                        <div className="flex justify-end gap-1">
                          {doc.status === 'draft' && (
                            <Button
                              variant="outline"
                              size="sm"
                              disabled={confirmingId === doc.id}
                              onClick={() => handleStart(doc)}
                            >
                              В работу
                            </Button>
                          )}
                          {doc.status === 'in_progress' && (
                            <Button
                              variant="outline"
                              size="sm"
                              disabled={confirmingId === doc.id}
                              onClick={() => handleClose(doc)}
                            >
                              <Check className="mr-1 h-3.5 w-3.5" />
                              Закрыть
                            </Button>
                          )}
                          <Button variant="ghost" size="sm" asChild>
                            <Link to={`/movement-documents/${doc.id}`}>Открыть</Link>
                          </Button>
                        </div>
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
