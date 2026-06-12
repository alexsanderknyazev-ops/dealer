import { useEffect, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { Pencil, Package } from 'lucide-react'
import * as api from './workOrdersApi'
import * as worksApi from './worksApi'
import * as partsApi from './partsApi'
import * as brandsApi from './brandsApi'
import * as vehiclesApi from './vehiclesApi'
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

function fmtTime(ts: number) {
  return ts ? new Date(ts * 1000).toLocaleString('ru-RU') : '—'
}

export function WorkOrderView() {
  const { id } = useParams()
  const navigate = useNavigate()
  const { logout, user } = useAuth()
  const [wo, setWo] = useState<api.WorkOrder | null>(null)
  const [works, setWorks] = useState<worksApi.Work[]>([])
  const [parts, setParts] = useState<partsApi.Part[]>([])
  const [laborRate, setLaborRate] = useState<brandsApi.ResolvedLaborRate | null>(null)
  const [vehicleBrandId, setVehicleBrandId] = useState('')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [issuing, setIssuing] = useState(false)

  function load() {
    if (!id) return
    setLoading(true)
    setError(null)
    api
      .getWorkOrder(id)
      .then(setWo)
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
    worksApi.listWorks({ limit: 1000 }).then((r) => setWorks(r.works)).catch(() => {})
    partsApi.listParts({ limit: 1000 }).then((r) => setParts(r.parts)).catch(() => {})
  }, [])

  useEffect(() => {
    if (!wo?.vehicle_id) return
    vehiclesApi.getVehicle(wo.vehicle_id).then((v) => setVehicleBrandId(v.brand_id || '')).catch(() => {})
  }, [wo?.vehicle_id])

  useEffect(() => {
    if (!vehicleBrandId || !wo?.dealer_point_id) {
      setLaborRate(null)
      return
    }
    brandsApi
      .resolveBrandLaborRate({
        brand_id: vehicleBrandId,
        dealer_point_id: wo.dealer_point_id,
        repair_type: wo.repair_type,
      })
      .then(setLaborRate)
      .catch(() => setLaborRate(null))
  }, [vehicleBrandId, wo?.dealer_point_id, wo?.repair_type])

  function workLabel(l: api.WorkOrderLabor) {
    if (l.work_name) return l.work_name
    if (l.description) return l.description
    const w = works.find((x) => x.id === l.work_id)
    return w ? `${w.code} — ${w.name}` : l.work_id || '—'
  }

  function workCode(l: api.WorkOrderLabor) {
    if (l.work_code) return l.work_code
    return works.find((x) => x.id === l.work_id)?.code || '—'
  }

  function workNormHours(l: api.WorkOrderLabor) {
    if (l.labor_hours) return l.labor_hours
    return works.find((x) => x.id === l.work_id)?.labor_hours || '—'
  }

  function partLabel(p: api.WorkOrderPart) {
    if (p.part_name) return p.part_name
    if (p.description) return p.description
    const part = parts.find((x) => x.id === p.part_id)
    return part ? `${part.sku} — ${part.name}` : p.part_id || '—'
  }

  function partSku(p: api.WorkOrderPart) {
    if (p.part_sku) return p.part_sku
    return parts.find((x) => x.id === p.part_id)?.sku || '—'
  }

  async function handleCreateMovementDoc() {
    if (!id || !wo) return
    setIssuing(true)
    setError(null)
    try {
      const updated = await api.movePartsToWork(id, user?.userId)
      setWo(updated)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Не удалось создать документ перемещения')
    } finally {
      setIssuing(false)
    }
  }

  const hasUnissuedParts = wo?.parts?.some((p) => !p.issued) ?? false
  const canCreateMovementDoc = hasUnissuedParts && !wo?.movement_document_id
  const movementDocActive =
    !!wo?.movement_document_id &&
    (wo.movement_document_status === 'draft' || wo.movement_document_status === 'in_progress')

  if (loading) return <LoadingState />
  if (!wo) return error ? <ErrorAlert message={error} onRetry={load} /> : null

  return (
    <div className="mx-auto w-full max-w-5xl space-y-6">
      <PageHeader
        title={wo.order_number}
        subtitle={
          <span className="flex flex-wrap items-center gap-2">
            <Badge>{STATUS_LABEL[wo.status] || wo.status}</Badge>
            <span className="text-muted-foreground">{REPAIR_LABEL[wo.repair_type] || wo.repair_type}</span>
          </span>
        }
        action={
          <div className="flex gap-2">
            {canCreateMovementDoc && (
              <Button variant="default" onClick={handleCreateMovementDoc} disabled={issuing}>
                <Package className="mr-2 h-4 w-4" />
                {issuing ? 'Создание…' : 'Создать документ перемещения'}
              </Button>
            )}
            {movementDocActive && (
              <Button variant="default" asChild>
                <Link to={`/movement-documents/${wo.movement_document_id}`}>
                  {wo.movement_document_status === 'in_progress' ? 'Закрыть перемещение' : 'Оформить перемещение'}
                </Link>
              </Button>
            )}
            <Button variant="outline" asChild>
              <Link to={`/work-orders/${wo.id}/edit`}>
                <Pencil className="mr-2 h-4 w-4" />
                Редактировать
              </Link>
            </Button>
          </div>
        }
      />

      {error && <ErrorAlert message={error} onRetry={load} />}

      <div className="grid gap-4 md:grid-cols-2">
        <Card>
          <CardHeader><CardTitle className="text-base">Основное</CardTitle></CardHeader>
          <CardContent className="space-y-2 text-sm">
            <div>Клиент: {wo.customer_name || wo.customer_id || '—'}</div>
            <div>
              Автомобиль: {wo.vehicle_label || wo.vehicle_id || '—'}
              {wo.vehicle_vin ? ` (VIN: ${wo.vehicle_vin})` : ''}
            </div>
            <div>Мастер-консультант: {wo.service_advisor_name || wo.service_advisor_id || '—'}</div>
            <div>Пробег: {wo.mileage_km} км</div>
            <div>Открыт: {fmtTime(wo.opened_at)}</div>
            <div>Закрыт: {fmtTime(wo.closed_at)}</div>
            <div>Документ перемещения: {wo.movement_document_id ? (
              <Link className="text-primary underline" to={`/movement-documents/${wo.movement_document_id}`}>
                {wo.movement_document_status || 'draft'}
              </Link>
            ) : 'не создан'}</div>
            <div>Запчасти списаны: {wo.parts_issued ? fmtTime(wo.parts_issued_at) : 'нет'}</div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader><CardTitle className="text-base">Стоимость</CardTitle></CardHeader>
          <CardContent className="space-y-2 text-sm">
            <div>Работы: {wo.labor_cost}</div>
            <div>Запчасти: {wo.parts_cost}</div>
            <div className="font-semibold">Итого: {wo.total_cost}</div>
            {laborRate?.found && (
              <div className="pt-2 text-muted-foreground">
                Н/ч ({REPAIR_LABEL[wo.repair_type] || wo.repair_type}): {laborRate.hour_price} ₽
                <span className="block text-xs">
                  Гарантия: {laborRate.warranty_hour_price} ₽ · Коммерция: {laborRate.commercial_hour_price} ₽
                </span>
              </div>
            )}
            {wo.complaint && <div className="pt-2">Жалоба: {wo.complaint}</div>}
            {wo.diagnosis && <div>Диагноз: {wo.diagnosis}</div>}
            {wo.notes && <div>Примечания: {wo.notes}</div>}
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader><CardTitle className="text-base">Работы</CardTitle></CardHeader>
        <CardContent className="p-0">
          {wo.labor?.length ? (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Код</TableHead>
                  <TableHead>Работа</TableHead>
                  <TableHead>Норма-час</TableHead>
                  <TableHead>Кол-во н/ч</TableHead>
                  <TableHead>Цена</TableHead>
                  <TableHead>Сумма</TableHead>
                  <TableHead>Исполнитель</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {wo.labor.map((l) => (
                  <TableRow key={l.id || l.work_id || l.description}>
                    <TableCell className="font-mono text-xs">{workCode(l)}</TableCell>
                    <TableCell>{workLabel(l)}</TableCell>
                    <TableCell>{workNormHours(l)}</TableCell>
                    <TableCell>{l.quantity}</TableCell>
                    <TableCell>{l.unit_price}</TableCell>
                    <TableCell>{l.amount}</TableCell>
                    <TableCell>{l.executor_name || l.executor_id || '—'}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          ) : (
            <p className="px-6 py-4 text-sm text-muted-foreground">Работы не добавлены</p>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader><CardTitle className="text-base">Запчасти</CardTitle></CardHeader>
        <CardContent className="p-0">
          {wo.parts?.length ? (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Артикул</TableHead>
                  <TableHead>Запчасть</TableHead>
                  <TableHead>Склад</TableHead>
                  <TableHead>Кол-во</TableHead>
                  <TableHead>Цена</TableHead>
                  <TableHead>Сумма</TableHead>
                  <TableHead>Статус</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {wo.parts.map((p) => (
                  <TableRow key={p.id || p.part_id}>
                    <TableCell className="font-mono text-xs">{partSku(p)}</TableCell>
                    <TableCell>{partLabel(p)}</TableCell>
                    <TableCell>{p.warehouse_name || p.warehouse_id}</TableCell>
                    <TableCell>{p.quantity}</TableCell>
                    <TableCell>{p.unit_price}</TableCell>
                    <TableCell>{p.amount}</TableCell>
                    <TableCell>{p.issued ? 'Списано' : 'В резерве'}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          ) : (
            <p className="px-6 py-4 text-sm text-muted-foreground">Запчасти не добавлены</p>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
