import { useEffect, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { ArrowLeft, Check, Pencil, Play, X } from 'lucide-react'
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
  sale: 'Реализация товара',
  receipt: 'Поступление товара',
}

const DEST_LABEL: Record<string, string> = {
  transfer: 'Другой склад',
  to_production: 'Производство',
  work_order_issue: 'В работу',
  from_production: 'Склад',
  receipt: 'Склад',
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

  async function handleStart() {
    if (!id) return
    setActing(true)
    setError(null)
    try {
      setDoc(await api.startMovementDocument(id))
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Не удалось перевести в работу')
    } finally {
      setActing(false)
    }
  }

  async function handleCreateExtraction() {
    if (!id) return
    setActing(true)
    setError(null)
    try {
      const created = await api.createProductionExtraction(id, user?.userId)
      navigate(`/movement-documents/${created.id}`)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Не удалось создать извлечение')
    } finally {
      setActing(false)
    }
  }

  async function handleClose() {
    if (!id || !doc) return
    const fromProduction = doc.movement_type === 'from_production'
    const isReceipt = doc.movement_type === 'receipt'
    const closeMsg = fromProduction
      ? 'Закрыть документ? Запчасти будут возвращены на склад, отменить операцию будет нельзя.'
      : isReceipt
        ? 'Закрыть документ? Запчасти поступят на склад, отменить операцию будет нельзя.'
        : 'Закрыть документ? Запчасти будут списаны со склада, отменить операцию будет нельзя.'
    if (!window.confirm(closeMsg)) {
      return
    }
    setActing(true)
    setError(null)
    try {
      const updated = await api.closeMovementDocument(id, user?.userId)
      setDoc(updated)
      if (updated.reference_type === 'work_order' && updated.reference_id) {
        navigate(`/work-orders/${updated.reference_id}`)
      }
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Не удалось закрыть документ')
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

  const canStart = doc.status === 'draft'
  const canClose = doc.status === 'in_progress'
  const canCancel = doc.status === 'draft' || doc.status === 'in_progress'
  const canEdit = doc.status === 'draft' || doc.status === 'in_progress'
  const canCreateExtraction =
    doc.status === 'closed' &&
    doc.movement_type !== 'from_production' &&
    ['work_order_issue', 'transfer', 'to_production'].includes(doc.movement_type)
  const closeLabel =
    doc.movement_type === 'from_production'
      ? 'Закрыть (вернуть на склад)'
      : doc.movement_type === 'receipt'
        ? 'Закрыть (оприходовать на склад)'
        : 'Закрыть (списать со склада)'

  return (
    <div className="mx-auto w-full max-w-5xl space-y-6">
      <PageHeader
        title={`Документ ${doc.document_number}`}
        subtitle={
          <div className="flex items-center gap-2">
            <Badge variant="secondary">{STATUS_LABEL[doc.status] || doc.status}</Badge>
            <Link className="text-sm text-muted-foreground underline" to="/movement-documents">
              Все документы
            </Link>
          </div>
        }
        action={
          canEdit || canStart || canClose || canCancel || canCreateExtraction ? (
            <div className="flex flex-wrap gap-2">
              {canCreateExtraction && (
                <Button onClick={handleCreateExtraction} disabled={acting}>
                  <ArrowLeft className="mr-2 h-4 w-4" />
                  Извлечение (вернуть на склад)
                </Button>
              )}
              {canEdit && (
                <Button variant="outline" asChild>
                  <Link to={`/movement-documents/${doc.id}/edit`}>
                    <Pencil className="mr-2 h-4 w-4" />
                    Редактировать
                  </Link>
                </Button>
              )}
              {canStart && (
                <Button onClick={handleStart} disabled={acting}>
                  <Play className="mr-2 h-4 w-4" />
                  В работу
                </Button>
              )}
              {canClose && (
                <Button onClick={handleClose} disabled={acting}>
                  <Check className="mr-2 h-4 w-4" />
                  {closeLabel}
                </Button>
              )}
              {canCancel && (
                <Button variant="outline" onClick={handleCancel} disabled={acting}>
                  <X className="mr-2 h-4 w-4" />
                  Отменить
                </Button>
              )}
            </div>
          ) : null
        }
      />

      {error && <ErrorAlert message={error} onRetry={load} />}

      <Card>
        <CardHeader><CardTitle className="text-base">Реквизиты</CardTitle></CardHeader>
        <CardContent className="space-y-2 text-sm">
          <div>Тип: {MOVEMENT_TYPE_LABEL[doc.movement_type] || doc.movement_type}</div>
          {doc.parent_document_id && (
            <div>
              Перемещение в производство:{' '}
              <Link className="text-primary underline" to={`/movement-documents/${doc.parent_document_id}`}>
                {doc.parent_document_number || doc.parent_document_id}
              </Link>
            </div>
          )}
          {doc.movement_type === 'receipt' && doc.supplier_name && <div>Поставщик: {doc.supplier_name}</div>}
          {doc.movement_type === 'receipt' && doc.receipt_warehouse_name && (
            <div>Склад поступления: {doc.receipt_warehouse_name}</div>
          )}
          {doc.movement_type === 'sale' && doc.customer_name && <div>Клиент: {doc.customer_name}</div>}
          {doc.movement_type === 'sale' && (doc.vehicle_vin || doc.vehicle_label) && (
            <div>
              Автомобиль: {doc.vehicle_label || '—'}
              {doc.vehicle_vin ? ` (VIN: ${doc.vehicle_vin})` : ''}
            </div>
          )}
          {doc.reference_type === 'work_order' && doc.reference_id && (
            <div>
              Заказ-наряд:{' '}
              <Link className="text-primary underline" to={`/work-orders/${doc.reference_id}`}>
                {doc.reference_label || 'открыть'}
              </Link>
              {doc.customer_name && <div>Клиент: {doc.customer_name}</div>}
              {(doc.vehicle_vin || doc.vehicle_label) && (
                <div>
                  Автомобиль: {doc.vehicle_label}
                  {doc.vehicle_vin ? ` (VIN: ${doc.vehicle_vin})` : ''}
                </div>
              )}
            </div>
          )}
          {doc.reference_type === 'supplier_order' && doc.reference_id && (
            <div>
              Заказ поставщику:{' '}
              <Link className="text-primary underline" to={`/supplier-orders/${doc.reference_id}`}>
                {doc.reference_label || doc.reference_id}
              </Link>
            </div>
          )}
          {doc.reference_type === 'customer_order' && doc.reference_id && (
            <div>
              Заказ покупателя:{' '}
              <Link className="text-primary underline" to={`/customer-orders/${doc.reference_id}`}>
                {doc.reference_label || doc.reference_id}
              </Link>
            </div>
          )}
          {doc.reference_type === 'movement_document' && doc.reference_id && !doc.parent_document_id && (
            <div>
              Основание:{' '}
              <Link className="text-primary underline" to={`/movement-documents/${doc.reference_id}`}>
                {doc.reference_label || doc.reference_id}
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
              Закрыл: {doc.confirmed_by_name || '—'} ({new Date(doc.confirmed_at * 1000).toLocaleString('ru-RU')})
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
                      <TableHead>Артикул</TableHead>
                      {doc.movement_type !== 'receipt' && <TableHead>Откуда</TableHead>}
                      {doc.movement_type !== 'receipt' && <TableHead>Куда</TableHead>}
                      {doc.movement_type === 'receipt' && <TableHead>Склад</TableHead>}
                      <TableHead>Кол-во</TableHead>
                      {doc.movement_type === 'receipt' ? (
                        <TableHead>Входная цена</TableHead>
                      ) : (
                        <TableHead>Остаток до списания</TableHead>
                      )}
                    </TableRow>
            </TableHeader>
            <TableBody>
              {doc.lines?.map((line) => (
                <TableRow key={line.id}>
                  <TableCell>{line.part_name || line.part_id}</TableCell>
                  <TableCell className="text-muted-foreground">{line.part_sku || '—'}</TableCell>
                  {doc.movement_type !== 'receipt' && (
                    <TableCell>{line.warehouse_name || line.warehouse_id}</TableCell>
                  )}
                  {doc.movement_type !== 'receipt' && (
                    <TableCell>
                      {line.destination_warehouse_name ||
                        line.destination_warehouse_id ||
                        DEST_LABEL[doc.movement_type] ||
                        '—'}
                    </TableCell>
                  )}
                  {doc.movement_type === 'receipt' && (
                    <TableCell>{doc.receipt_warehouse_name || line.warehouse_name || '—'}</TableCell>
                  )}
                  <TableCell>{line.quantity}</TableCell>
                  {doc.movement_type === 'receipt' ? (
                    <TableCell>{line.unit_cost || '—'}</TableCell>
                  ) : (
                    <TableCell className="text-muted-foreground">
                      {line.source_stock_quantity > 0 ? line.source_stock_quantity : '—'}
                    </TableCell>
                  )}
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  )
}
