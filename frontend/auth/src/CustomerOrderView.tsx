import { useEffect, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { ArrowLeft, Pencil, ShoppingCart, Truck, Wrench } from 'lucide-react'
import * as api from './customerOrdersApi'
import * as vehiclesApi from './vehiclesApi'
import { useAuth } from './auth'
import { PageHeader } from '@/components/common/PageHeader'
import { ErrorAlert } from '@/components/common/ErrorAlert'
import { LoadingState } from '@/components/common/LoadingState'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { FormField } from '@/components/common/FormField'
import { NativeSelect } from '@/components/ui/native-select'
import { Input } from '@/components/ui/input'

const STATUS_LABEL: Record<string, string> = {
  draft: 'Черновик',
  linked: 'Связан с исполнением',
  fulfilled: 'Выполнен',
  cancelled: 'Отменён',
}

export function CustomerOrderView() {
  const { id } = useParams()
  const navigate = useNavigate()
  const { logout, user } = useAuth()
  const [order, setOrder] = useState<api.CustomerOrder | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [acting, setActing] = useState(false)
  const [showWoForm, setShowWoForm] = useState(false)
  const [vehicles, setVehicles] = useState<vehiclesApi.Vehicle[]>([])
  const [woVehicleId, setWoVehicleId] = useState('')
  const [woVehicleVin, setWoVehicleVin] = useState('')

  function load() {
    if (!id) return
    setLoading(true)
    api
      .getCustomerOrder(id)
      .then(setOrder)
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

  useEffect(() => {
    vehiclesApi.listVehicles({ limit: 500 }).then((r) => setVehicles(r.vehicles)).catch(() => {})
  }, [])

  async function handleCreateSale() {
    if (!id) return
    setActing(true)
    try {
      const doc = await api.createSaleFromCustomerOrder(id, user?.userId)
      navigate(`/movement-documents/${doc.id}`)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Не удалось создать реализацию')
    } finally {
      setActing(false)
    }
  }

  async function handleCreateWorkOrder() {
    if (!id) return
    if (!order?.vehicle_id && !woVehicleId && !woVehicleVin.trim()) {
      setError('Укажите автомобиль или VIN для заказ-наряда')
      return
    }
    setActing(true)
    try {
      const wo = await api.createWorkOrderFromCustomerOrder(id, {
        vehicle_id: woVehicleId || undefined,
        vehicle_vin: woVehicleVin.trim() || undefined,
      })
      navigate(`/work-orders/${wo.work_order_id}`)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Не удалось создать заказ-наряд')
    } finally {
      setActing(false)
    }
  }

  if (loading) return <LoadingState />
  if (!order) return error ? <ErrorAlert message={error} onRetry={load} /> : null

  const canEdit = order.status === 'draft'
  const canFulfill = order.status === 'draft' && !order.fulfillment_movement_document_id && !order.fulfillment_work_order_id
  const needsVehicleForWo = !order.vehicle_id

  return (
    <div className="mx-auto w-full max-w-5xl space-y-6">
      <PageHeader
        title={`Заказ ${order.order_number}`}
        subtitle={
          <div className="flex items-center gap-2">
            <Badge variant="secondary">{STATUS_LABEL[order.status] || order.status}</Badge>
            <Link className="text-sm text-muted-foreground underline" to="/customer-orders">Все заказы</Link>
          </div>
        }
        actions={
          <div className="flex flex-wrap gap-2">
            {canFulfill && (
              <>
                <Button onClick={handleCreateSale} disabled={acting}>
                  <ShoppingCart className="mr-1 h-4 w-4" /> Создать реализацию
                </Button>
                <Button
                  variant="secondary"
                  onClick={() => {
                    if (needsVehicleForWo) {
                      setShowWoForm(true)
                    } else {
                      handleCreateWorkOrder()
                    }
                  }}
                  disabled={acting}
                >
                  <Wrench className="mr-1 h-4 w-4" /> Создать заказ-наряд
                </Button>
              </>
            )}
            {canEdit && (
              <>
                <Button variant="outline" asChild>
                  <Link to={`/supplier-orders/new?customer_order_id=${order.id}`}>
                    <Truck className="mr-1 h-4 w-4" /> Заказать у поставщика
                  </Link>
                </Button>
                <Button variant="outline" asChild>
                  <Link to={`/customer-orders/${order.id}/edit`}><Pencil className="mr-1 h-4 w-4" /> Редактировать</Link>
                </Button>
              </>
            )}
          </div>
        }
      />
      {error && <ErrorAlert message={error} onRetry={load} />}
      {showWoForm && canFulfill && needsVehicleForWo && (
        <Card>
          <CardHeader><CardTitle className="text-base">Автомобиль для заказ-наряда</CardTitle></CardHeader>
          <CardContent className="grid gap-4 sm:grid-cols-2">
            <FormField label="Автомобиль" htmlFor="wo_vehicle_id">
              <NativeSelect id="wo_vehicle_id" value={woVehicleId} onChange={(e) => setWoVehicleId(e.target.value)}>
                <option value="">— выберите автомобиль —</option>
                {vehicles.map((v) => (
                  <option key={v.id} value={v.id}>{v.make} {v.model} {v.year}{v.vin ? ` (${v.vin})` : ''}</option>
                ))}
              </NativeSelect>
            </FormField>
            <FormField label="или VIN" htmlFor="wo_vehicle_vin">
              <Input id="wo_vehicle_vin" value={woVehicleVin} onChange={(e) => setWoVehicleVin(e.target.value)} placeholder="VIN" />
            </FormField>
            <div className="flex items-end gap-2">
              <Button onClick={handleCreateWorkOrder} disabled={acting}>Создать с запчастями из заказа</Button>
              <Button variant="ghost" onClick={() => setShowWoForm(false)}>Отмена</Button>
            </div>
          </CardContent>
        </Card>
      )}
      <Card>
        <CardHeader><CardTitle className="text-base">Реквизиты</CardTitle></CardHeader>
        <CardContent className="space-y-2 text-sm">
          <div>Клиент: {order.customer_name}</div>
          <div>Склад отгрузки: {order.issue_warehouse_name}</div>
          {(order.vehicle_vin || order.vehicle_label) && (
            <div>Автомобиль: {order.vehicle_label || '—'}{order.vehicle_vin ? ` (VIN: ${order.vehicle_vin})` : ''}</div>
          )}
          {order.fulfillment_movement_document_id && (
            <div>
              Документ реализации:{' '}
              <Link className="text-primary underline" to={`/movement-documents/${order.fulfillment_movement_document_id}`}>
                {order.fulfillment_movement_document_number || order.fulfillment_movement_document_id}
              </Link>
            </div>
          )}
          {order.fulfillment_work_order_id && (
            <div>
              Заказ-наряд:{' '}
              <Link className="text-primary underline" to={`/work-orders/${order.fulfillment_work_order_id}`}>
                {order.fulfillment_work_order_number || order.fulfillment_work_order_id}
              </Link>
            </div>
          )}
          {order.notes && <div>Примечание: {order.notes}</div>}
        </CardContent>
      </Card>
      <Card>
        <CardHeader><CardTitle className="text-base">Строки заказа</CardTitle></CardHeader>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Запчасть</TableHead>
                <TableHead>Артикул</TableHead>
                <TableHead>Кол-во</TableHead>
                <TableHead>Цена</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {order.lines?.map((l) => (
                <TableRow key={l.id}>
                  <TableCell>{l.part_name || l.part_id}</TableCell>
                  <TableCell>{l.part_sku || '—'}</TableCell>
                  <TableCell>{l.quantity}</TableCell>
                  <TableCell>{l.unit_price}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
      <Button variant="ghost" asChild>
        <Link to="/customer-orders"><ArrowLeft className="mr-1 h-4 w-4" /> К списку</Link>
      </Button>
    </div>
  )
}
