import { useEffect, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { ArrowLeft, Pencil, Truck, Wrench } from 'lucide-react'
import * as api from './supplierOrdersApi'
import * as customersApi from './customersApi'
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

export function SupplierOrderView() {
  const { id } = useParams()
  const navigate = useNavigate()
  const { logout, user } = useAuth()
  const [order, setOrder] = useState<api.SupplierOrder | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [acting, setActing] = useState(false)
  const [showWoForm, setShowWoForm] = useState(false)
  const [customers, setCustomers] = useState<customersApi.Customer[]>([])
  const [vehicles, setVehicles] = useState<vehiclesApi.Vehicle[]>([])
  const [woCustomerId, setWoCustomerId] = useState('')
  const [woVehicleId, setWoVehicleId] = useState('')
  const [woVehicleVin, setWoVehicleVin] = useState('')

  function load() {
    if (!id) return
    setLoading(true)
    api
      .getSupplierOrder(id)
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
    customersApi.listCustomers({ limit: 500 }).then((r) => setCustomers(r.customers)).catch(() => {})
    vehiclesApi.listVehicles({ limit: 500 }).then((r) => setVehicles(r.vehicles)).catch(() => {})
  }, [])

  async function handleCreateReceipt() {
    if (!id) return
    setActing(true)
    try {
      const doc = await api.createReceiptFromSupplierOrder(id, user?.userId)
      navigate(`/movement-documents/${doc.id}`)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Не удалось создать поступление')
    } finally {
      setActing(false)
    }
  }

  async function handleCreateWorkOrder() {
    if (!id || !woCustomerId) {
      setError('Укажите клиента для заказ-наряда')
      return
    }
    if (!woVehicleId && !woVehicleVin.trim()) {
      setError('Укажите автомобиль или VIN для заказ-наряда')
      return
    }
    setActing(true)
    try {
      const wo = await api.createWorkOrderFromSupplierOrder(id, {
        customer_id: woCustomerId,
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

  return (
    <div className="mx-auto w-full max-w-5xl space-y-6">
      <PageHeader
        title={`Заказ ${order.order_number}`}
        subtitle={
          <div className="flex items-center gap-2">
            <Badge variant="secondary">{STATUS_LABEL[order.status] || order.status}</Badge>
            <Link className="text-sm text-muted-foreground underline" to="/supplier-orders">Все заказы</Link>
          </div>
        }
        action={
          <div className="flex flex-wrap gap-2">
            {canFulfill && (
              <>
                <Button onClick={handleCreateReceipt} disabled={acting}>
                  <Truck className="mr-1 h-4 w-4" /> Создать поступление
                </Button>
                <Button variant="secondary" onClick={() => setShowWoForm((v) => !v)} disabled={acting}>
                  <Wrench className="mr-1 h-4 w-4" /> Создать заказ-наряд
                </Button>
              </>
            )}
            {canEdit && (
              <Button variant="outline" asChild>
                <Link to={`/supplier-orders/${order.id}/edit`}><Pencil className="mr-1 h-4 w-4" /> Редактировать</Link>
              </Button>
            )}
          </div>
        }
      />
      {error && <ErrorAlert message={error} onRetry={load} />}
      {showWoForm && canFulfill && (
        <Card>
          <CardHeader><CardTitle className="text-base">Новый заказ-наряд</CardTitle></CardHeader>
          <CardContent className="grid gap-4 sm:grid-cols-2">
            <FormField label="Клиент" htmlFor="wo_customer_id">
              <NativeSelect id="wo_customer_id" value={woCustomerId} onChange={(e) => setWoCustomerId(e.target.value)}>
                <option value="">— выберите клиента —</option>
                {customers.map((c) => (
                  <option key={c.id} value={c.id}>{c.name}</option>
                ))}
              </NativeSelect>
            </FormField>
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
            <div className="flex items-end">
              <Button onClick={handleCreateWorkOrder} disabled={acting}>Создать с запчастями из заказа</Button>
            </div>
          </CardContent>
        </Card>
      )}
      <Card>
        <CardHeader><CardTitle className="text-base">Реквизиты</CardTitle></CardHeader>
        <CardContent className="space-y-2 text-sm">
          <div>Поставщик: {order.supplier_name}</div>
          <div>Склад поступления: {order.receipt_warehouse_name}</div>
          {order.fulfillment_movement_document_id && (
            <div>
              Документ поступления:{' '}
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
        <Link to="/supplier-orders"><ArrowLeft className="mr-1 h-4 w-4" /> К списку</Link>
      </Button>
    </div>
  )
}
