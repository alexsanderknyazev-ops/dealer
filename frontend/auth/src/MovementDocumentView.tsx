import { useEffect, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { Check, X } from 'lucide-react'
import * as api from './movementDocumentsApi'
import { useAuth } from './auth'
import { PageHeader } from '@/components/common/PageHeader'
import { ErrorAlert } from '@/components/common/ErrorAlert'
import { LoadingState } from '@/components/common/LoadingState'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'

const STATUS_LABEL: Record<string, string> = {
  draft: 'Черновик',
  confirmed: 'Подтверждён',
  cancelled: 'Отменён',
}

export function MovementDocumentView() {
  const { id } = useParams()
  const navigate = useNavigate()
  const { logout, user } = useAuth()
  const [doc, setDoc] = useState<api.MovementDocument | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [acting, setActing] = useState(false)

  function load() {
    if (!id) return
    setLoading(true)
    setError(null)
    api
      .getMovementDocument(id)
      .then(setDoc)
      .catch(async (e) => {
        if (e instanceof api.ApiError && (e.status === 401 || e.status === 403)) {
          await logout()
          navigate('/login', { replace: true })
          return
        }
        setError(e instanceof Error ? e.message : 'Ошибка загрузки')
      })
      .finally(() => setLoading(false))
  }

  useEffect(() => {
    load()
  }, [id])

  async function handleConfirm() {
    if (!id) return
    setActing(true)
    setError(null)
    try {
      const updated = await api.confirmMovementDocument(id, user?.userId)
      setDoc(updated)
      if (updated.reference_type === 'work_order' && updated.reference_id) {
        navigate(`/work-orders/${updated.reference_id}`)
      }
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Не удалось подтвердить документ')
    } finally {
      setActing(false)
    }
  }

  async function handleCancel() {
    if (!id) return
    setActing(true)
    setError(null)
    try {
      setDoc(await api.cancelMovementDocument(id))
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Не удалось отменить документ')
    } finally {
      setActing(false)
    }
  }

  if (loading) return <LoadingState />
  if (!doc) return error ? <ErrorAlert message={error} onRetry={load} /> : null

  return (
    <div className="mx-auto w-full max-w-5xl space-y-6">
      <PageHeader
        title={`Документ ${doc.document_number}`}
        subtitle={<Badge variant="secondary">{STATUS_LABEL[doc.status] || doc.status}</Badge>}
        action={
          doc.status === 'draft' ? (
            <div className="flex gap-2">
              <Button onClick={handleConfirm} disabled={acting}>
                <Check className="mr-2 h-4 w-4" />
                Подтвердить перемещение
              </Button>
              <Button variant="outline" onClick={handleCancel} disabled={acting}>
                <X className="mr-2 h-4 w-4" />
                Отменить
              </Button>
            </div>
          ) : null
        }
      />

      {error && <ErrorAlert message={error} onRetry={load} />}

      <Card>
        <CardHeader><CardTitle className="text-base">Реквизиты</CardTitle></CardHeader>
        <CardContent className="space-y-2 text-sm">
          <div>Тип: {doc.movement_type}</div>
          {doc.reference_type === 'work_order' && doc.reference_id && (
            <div>
              Заказ-наряд:{' '}
              <Link className="text-primary underline" to={`/work-orders/${doc.reference_id}`}>
                открыть
              </Link>
            </div>
          )}
          {doc.notes && <div>Примечание: {doc.notes}</div>}
          <div>
            Создал: {doc.created_by_name || '—'}
            {doc.created_at > 0 && ` (${new Date(doc.created_at * 1000).toLocaleString('ru-RU')})`}
          </div>
          {doc.confirmed_at > 0 && (
            <div>
              Подтвердил: {doc.confirmed_by_name || '—'} ({new Date(doc.confirmed_at * 1000).toLocaleString('ru-RU')})
            </div>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader><CardTitle className="text-base">Строки перемещения</CardTitle></CardHeader>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Запчасть</TableHead>
                <TableHead>Склад</TableHead>
                <TableHead>Кол-во</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {doc.lines?.map((line) => (
                <TableRow key={line.id}>
                  <TableCell className="font-mono text-xs">{line.part_id}</TableCell>
                  <TableCell className="font-mono text-xs">{line.warehouse_id}</TableCell>
                  <TableCell>{line.quantity}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  )
}
