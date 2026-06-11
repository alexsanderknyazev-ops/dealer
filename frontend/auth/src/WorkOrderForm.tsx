import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { Plus, Trash2 } from 'lucide-react'
import type { WorkOrderForm as FormType, WorkOrderLabor, WorkOrderPart } from './workOrdersApi'
import * as api from './workOrdersApi'
import * as customersApi from './customersApi'
import * as vehiclesApi from './vehiclesApi'
import * as partsApi from './partsApi'
import * as worksApi from './worksApi'
import * as employeesApi from './employeesApi'
import * as dealerPointsApi from './dealerPointsApi'
import { useAuth } from './auth'
import { FormPage } from '@/components/common/FormPage'
import { FormField } from '@/components/common/FormField'
import { FormActions } from '@/components/common/FormActions'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Input } from '@/components/ui/input'
import { NativeSelect } from '@/components/ui/native-select'
import { Textarea } from '@/components/ui/textarea'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

const emptyLabor = (): WorkOrderLabor => ({ work_id: '', quantity: '1', unit_price: '0', executor_id: '' })
const emptyPart = (): WorkOrderPart => ({ part_id: '', warehouse_id: '', description: '', quantity: '1', unit_price: '0' })

function toUnix(dt: string): number | undefined {
  if (!dt) return undefined
  const ms = Date.parse(dt)
  return Number.isFinite(ms) ? Math.floor(ms / 1000) : undefined
}

function fromUnix(ts?: number): string {
  if (!ts) return ''
  const d = new Date(ts * 1000)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`
}

export function WorkOrderForm() {
  const { id } = useParams()
  const navigate = useNavigate()
  const { logout } = useAuth()
  const isNew = id === 'new' || !id
  const [loading, setLoading] = useState(!isNew)
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)
  const [customers, setCustomers] = useState<customersApi.Customer[]>([])
  const [vehicles, setVehicles] = useState<vehiclesApi.Vehicle[]>([])
  const [parts, setParts] = useState<partsApi.Part[]>([])
  const [works, setWorks] = useState<worksApi.Work[]>([])
  const [employees, setEmployees] = useState<employeesApi.Employee[]>([])
  const [warehouses, setWarehouses] = useState<dealerPointsApi.Warehouse[]>([])
  const [openedAt, setOpenedAt] = useState('')
  const [closedAt, setClosedAt] = useState('')
  const [form, setForm] = useState<FormType>({
    customer_id: '',
    vehicle_id: '',
    repair_type: 'commercial',
    status: 'draft',
    mileage_km: 0,
    labor: [emptyLabor()],
    parts: [],
  })

  useEffect(() => {
    customersApi.listCustomers({ limit: 500 }).then((r) => setCustomers(r.customers)).catch(() => {})
    vehiclesApi.listVehicles({ limit: 500 }).then((r) => setVehicles(r.vehicles)).catch(() => {})
    partsApi.listParts({ limit: 500 }).then((r) => setParts(r.parts)).catch(() => {})
    worksApi.listWorks({ limit: 500 }).then((r) => setWorks(r.works)).catch(() => {})
    employeesApi.listEmployees({ limit: 500, active_only: true }).then((r) => setEmployees(r.employees)).catch(() => {})
    dealerPointsApi.listWarehouses({ limit: 500 }).then((r) => setWarehouses(r.warehouses)).catch(() => {})
  }, [])

  useEffect(() => {
    if (isNew) return
    api
      .getWorkOrder(id!)
      .then((wo) => {
        setForm({
          customer_id: wo.customer_id,
          vehicle_id: wo.vehicle_id,
          dealer_point_id: wo.dealer_point_id,
          warehouse_id: wo.warehouse_id,
          repair_type: wo.repair_type,
          status: wo.status,
          service_advisor_id: wo.service_advisor_id,
          complaint: wo.complaint,
          diagnosis: wo.diagnosis,
          mileage_km: wo.mileage_km,
          notes: wo.notes,
          labor: wo.labor?.length
            ? wo.labor.map((l) => ({ ...l, work_id: l.work_id || '', executor_id: l.executor_id || '' }))
            : [emptyLabor()],
          parts: wo.parts?.filter((p) => !p.issued) || [],
        })
        setOpenedAt(fromUnix(wo.opened_at))
        setClosedAt(fromUnix(wo.closed_at))
      })
      .catch(async (e) => {
        if (e instanceof api.ApiError && (e.status === 401 || e.status === 403)) {
          await logout()
          navigate('/login', { replace: true })
          return
        }
        setError(e instanceof Error ? e.message : 'Ошибка загрузки')
      })
      .finally(() => setLoading(false))
  }, [id, isNew, logout, navigate])

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setSubmitting(true)
    setError(null)
    const payload: FormType = {
      ...form,
      opened_at: toUnix(openedAt),
      closed_at: toUnix(closedAt),
      labor: form.labor?.filter((l) => l.work_id),
      parts: form.parts?.filter((p) => p.part_id),
    }
    try {
      const saved = isNew ? await api.createWorkOrder(payload) : await api.updateWorkOrder(id!, payload)
      navigate(`/work-orders/${saved.id}`)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Ошибка сохранения')
    } finally {
      setSubmitting(false)
    }
  }

  function updateLabor(index: number, patch: Partial<WorkOrderLabor>) {
    setForm((f) => {
      const labor = [...(f.labor || [])]
      labor[index] = { ...labor[index], ...patch }
      return { ...f, labor }
    })
  }

  function updatePart(index: number, patch: Partial<WorkOrderPart>) {
    setForm((f) => {
      const partsRows = [...(f.parts || [])]
      partsRows[index] = { ...partsRows[index], ...patch }
      return { ...f, parts: partsRows }
    })
  }

  return (
    <FormPage title={isNew ? 'Новый заказ-наряд' : 'Редактирование заказ-наряда'} loading={loading}>
      <form onSubmit={handleSubmit} className="space-y-6">
        {error && (
          <Alert variant="destructive">
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}

        <div className="grid gap-4 md:grid-cols-2">
          <FormField label="Клиент" required>
            <NativeSelect
              value={form.customer_id}
              onChange={(e) => setForm({ ...form, customer_id: e.target.value })}
              required
            >
              <option value="">Выберите клиента</option>
              {customers.map((c) => (
                <option key={c.id} value={c.id}>{c.name || c.email}</option>
              ))}
            </NativeSelect>
          </FormField>
          <FormField label="Автомобиль" required>
            <NativeSelect
              value={form.vehicle_id}
              onChange={(e) => setForm({ ...form, vehicle_id: e.target.value })}
              required
            >
              <option value="">Выберите автомобиль</option>
              {vehicles.map((v) => (
                <option key={v.id} value={v.id}>{v.vin} — {v.make} {v.model}</option>
              ))}
            </NativeSelect>
          </FormField>
          <FormField label="Вид ремонта">
            <NativeSelect value={form.repair_type} onChange={(e) => setForm({ ...form, repair_type: e.target.value })}>
              <option value="warranty_manufacturer">Гарантия производителя</option>
              <option value="pre_sale">Предпродажная подготовка</option>
              <option value="commercial">Коммерческий ремонт</option>
              <option value="maintenance">Техобслуживание</option>
            </NativeSelect>
          </FormField>
          <FormField label="Статус">
            <NativeSelect value={form.status} onChange={(e) => setForm({ ...form, status: e.target.value })}>
              <option value="draft">Черновик</option>
              <option value="in_progress">В работе</option>
              <option value="completed">Выполнен</option>
              <option value="closed">Закрыт</option>
              <option value="paid">Оплачен</option>
            </NativeSelect>
          </FormField>
          <FormField label="Мастер-консультант">
            <NativeSelect
              value={form.service_advisor_id || ''}
              onChange={(e) => setForm({ ...form, service_advisor_id: e.target.value })}
            >
              <option value="">Не выбран</option>
              {employees.map((emp) => (
                <option key={emp.id} value={employeesApi.employeeRefId(emp)}>
                  {emp.full_name} — {emp.position}
                </option>
              ))}
            </NativeSelect>
          </FormField>
          <FormField label="Склад по умолчанию">
            <NativeSelect
              value={form.warehouse_id || ''}
              onChange={(e) => setForm({ ...form, warehouse_id: e.target.value })}
            >
              <option value="">Не выбран</option>
              {warehouses.map((w) => (
                <option key={w.id} value={w.id}>{w.name}</option>
              ))}
            </NativeSelect>
          </FormField>
          <FormField label="Время открытия">
            <Input type="datetime-local" value={openedAt} onChange={(e) => setOpenedAt(e.target.value)} />
          </FormField>
          <FormField label="Время закрытия">
            <Input type="datetime-local" value={closedAt} onChange={(e) => setClosedAt(e.target.value)} />
          </FormField>
          <FormField label="Пробег, км">
            <Input
              type="number"
              min={0}
              value={form.mileage_km ?? 0}
              onChange={(e) => setForm({ ...form, mileage_km: Number(e.target.value) })}
            />
          </FormField>
        </div>

        <FormField label="Жалоба клиента">
          <Textarea value={form.complaint || ''} onChange={(e) => setForm({ ...form, complaint: e.target.value })} />
        </FormField>
        <FormField label="Диагноз">
          <Textarea value={form.diagnosis || ''} onChange={(e) => setForm({ ...form, diagnosis: e.target.value })} />
        </FormField>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between">
            <CardTitle className="text-base">Работы</CardTitle>
            <Button type="button" variant="outline" size="sm" onClick={() => setForm((f) => ({ ...f, labor: [...(f.labor || []), emptyLabor()] }))}>
              <Plus className="mr-1 h-4 w-4" /> Добавить
            </Button>
          </CardHeader>
          <CardContent className="space-y-3">
            {(form.labor || []).map((l, i) => (
              <div key={i} className="grid gap-2 rounded-md border p-3 md:grid-cols-5">
                <NativeSelect
                  value={l.work_id}
                  onChange={(e) => {
                    const work = works.find((x) => x.id === e.target.value)
                    updateLabor(i, {
                      work_id: e.target.value,
                      description: work?.name || '',
                      quantity: work?.labor_hours || '1',
                      unit_price: work?.unit_price || '0',
                    })
                  }}
                >
                  <option value="">Работа</option>
                  {works.map((work) => (
                    <option key={work.id} value={work.id}>{work.code} — {work.name}</option>
                  ))}
                </NativeSelect>
                <Input placeholder="Нормо-часы" value={l.quantity} onChange={(e) => updateLabor(i, { quantity: e.target.value })} />
                <Input placeholder="Цена" value={l.unit_price} onChange={(e) => updateLabor(i, { unit_price: e.target.value })} />
                <NativeSelect
                  value={l.executor_id}
                  onChange={(e) => updateLabor(i, { executor_id: e.target.value })}
                >
                  <option value="">Исполнитель</option>
                  {employees.map((emp) => (
                    <option key={emp.id} value={employeesApi.employeeRefId(emp)}>
                      {emp.full_name}
                    </option>
                  ))}
                </NativeSelect>
                <Button type="button" variant="ghost" size="icon" onClick={() => setForm((f) => ({ ...f, labor: f.labor?.filter((_, idx) => idx !== i) }))}>
                  <Trash2 className="h-4 w-4" />
                </Button>
              </div>
            ))}
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between">
            <CardTitle className="text-base">Запчасти</CardTitle>
            <Button type="button" variant="outline" size="sm" onClick={() => setForm((f) => ({ ...f, parts: [...(f.parts || []), emptyPart()] }))}>
              <Plus className="mr-1 h-4 w-4" /> Добавить
            </Button>
          </CardHeader>
          <CardContent className="space-y-3">
            {(form.parts || []).map((p, i) => (
              <div key={i} className="grid gap-2 rounded-md border p-3 md:grid-cols-6">
                <NativeSelect value={p.part_id} onChange={(e) => {
                  const part = parts.find((x) => x.id === e.target.value)
                  updatePart(i, { part_id: e.target.value, description: part?.name || '', unit_price: part?.price || '0', warehouse_id: p.warehouse_id || part?.warehouse_id || form.warehouse_id || '' })
                }}>
                  <option value="">Запчасть</option>
                  {parts.map((part) => <option key={part.id} value={part.id}>{part.sku} — {part.name}</option>)}
                </NativeSelect>
                <NativeSelect value={p.warehouse_id} onChange={(e) => updatePart(i, { warehouse_id: e.target.value })}>
                  <option value="">Склад</option>
                  {warehouses.map((w) => <option key={w.id} value={w.id}>{w.name}</option>)}
                </NativeSelect>
                <Input placeholder="Кол-во" value={p.quantity} onChange={(e) => updatePart(i, { quantity: e.target.value })} />
                <Input placeholder="Цена" value={p.unit_price} onChange={(e) => updatePart(i, { unit_price: e.target.value })} />
                <Input placeholder="Описание" value={p.description} onChange={(e) => updatePart(i, { description: e.target.value })} />
                <Button type="button" variant="ghost" size="icon" onClick={() => setForm((f) => ({ ...f, parts: f.parts?.filter((_, idx) => idx !== i) }))}>
                  <Trash2 className="h-4 w-4" />
                </Button>
              </div>
            ))}
          </CardContent>
        </Card>

        <FormField label="Примечания">
          <Textarea value={form.notes || ''} onChange={(e) => setForm({ ...form, notes: e.target.value })} />
        </FormField>

        <FormActions
          submitting={submitting}
          submitLabel={isNew ? 'Создать' : 'Сохранить'}
          onCancel={() => navigate('/work-orders')}
        />
      </form>
    </FormPage>
  )
}
