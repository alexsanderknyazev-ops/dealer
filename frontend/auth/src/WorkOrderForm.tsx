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
import * as brandsApi from './brandsApi'
import * as dealsApi from './dealsApi'
import * as customerOrdersApi from './customerOrdersApi'
import { useAuth } from './auth'
import { FormPage } from '@/components/common/FormPage'
import { FormField } from '@/components/common/FormField'
import { FormActions } from '@/components/common/FormActions'
import { SearchSelect, type SearchSelectItem } from '@/components/common/SearchSelect'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Input } from '@/components/ui/input'
import { NativeSelect } from '@/components/ui/native-select'
import { Textarea } from '@/components/ui/textarea'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

type LaborRow = WorkOrderLabor & { display?: string }
type PartRow = WorkOrderPart & { display?: string }

type WorkOrderFormState = Omit<FormType, 'labor' | 'parts'> & { labor?: LaborRow[]; parts?: PartRow[] }

const emptyLabor = (): LaborRow => ({ work_id: '', description: '', quantity: '1', unit_price: '0', executor_id: '' })
const emptyPart = (): PartRow => ({ part_id: '', warehouse_id: '', description: '', quantity: '1', unit_price: '0' })

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

function customerItem(c: customersApi.Customer): SearchSelectItem<customersApi.Customer> {
  return {
    value: c.id,
    label: c.name || c.email,
    sublabel: [c.email, c.phone].filter(Boolean).join(' · '),
    data: c,
  }
}

function vehicleItem(v: vehiclesApi.Vehicle): SearchSelectItem<vehiclesApi.Vehicle> {
  return {
    value: v.id,
    label: `${v.make} ${v.model}${v.year ? ` ${v.year}` : ''}`,
    sublabel: v.vin,
    data: v,
  }
}

function partItem(p: partsApi.Part): SearchSelectItem<partsApi.Part> {
  return {
    value: p.id,
    label: p.sku,
    sublabel: p.name,
    data: p,
  }
}

function workItem(w: worksApi.Work): SearchSelectItem<worksApi.Work> {
  return {
    value: w.id,
    label: w.code,
    sublabel: w.name,
    data: w,
  }
}

export function WorkOrderForm() {
  const { id } = useParams()
  const navigate = useNavigate()
  const { logout } = useAuth()
  const isNew = id === 'new' || !id
  const [loading, setLoading] = useState(!isNew)
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)
  const [employees, setEmployees] = useState<employeesApi.Employee[]>([])
  const [warehouses, setWarehouses] = useState<dealerPointsApi.Warehouse[]>([])
  const [dealerPoints, setDealerPoints] = useState<dealerPointsApi.DealerPoint[]>([])
  const [customerDisplay, setCustomerDisplay] = useState('')
  const [vehicleDisplay, setVehicleDisplay] = useState('')
  const [customerVehicles, setCustomerVehicles] = useState<SearchSelectItem<vehiclesApi.Vehicle>[]>([])
  const [customerPresets, setCustomerPresets] = useState<SearchSelectItem<customersApi.Customer>[]>([])
  const [hourPrice, setHourPrice] = useState('')
  const [laborRateNote, setLaborRateNote] = useState('')
  const [openedAt, setOpenedAt] = useState('')
  const [closedAt, setClosedAt] = useState('')
  const [form, setForm] = useState<WorkOrderFormState>({
    customer_id: '',
    vehicle_id: '',
    repair_type: 'commercial',
    status: 'draft',
    mileage_km: 0,
    labor: [emptyLabor()],
    parts: [],
  })

  useEffect(() => {
    employeesApi.listEmployees({ limit: 500, active_only: true }).then((r) => setEmployees(r.employees)).catch(() => {})
    dealerPointsApi.listWarehouses({ limit: 500 }).then((r) => setWarehouses(r.warehouses)).catch(() => {})
    dealerPointsApi.listDealerPoints({ limit: 200 }).then((r) => setDealerPoints(r.dealer_points)).catch(() => {})
  }, [])

  async function loadCustomerVehicles(customerId: string) {
    if (!customerId) {
      setCustomerVehicles([])
      return
    }
    const result: SearchSelectItem<vehiclesApi.Vehicle>[] = []
    const seen = new Set<string>()
    const push = (vid: string, label: string, sublabel?: string) => {
      if (!vid || seen.has(vid)) return
      seen.add(vid)
      result.push({ value: vid, label, sublabel, data: undefined })
    }
    try {
      const [woRes, dealRes, coRes] = await Promise.all([
        api.listWorkOrders({ customer_id: customerId, limit: 100 }),
        dealsApi.listDeals({ customer_id: customerId, limit: 100 }),
        customerOrdersApi.listCustomerOrders({ limit: 200 }),
      ])
      for (const wo of woRes.work_orders || []) {
        push(wo.vehicle_id, wo.vehicle_label || wo.vehicle_vin, wo.vehicle_vin)
      }
      for (const d of dealRes.deals || []) {
        push(d.vehicle_id, 'Автомобиль по сделке', d.vehicle_id)
      }
      for (const o of coRes.orders || []) {
        if (o.customer_id === customerId) push(o.vehicle_id, o.vehicle_label || o.vehicle_vin, o.vehicle_vin)
      }
    } catch {
      // не критично — поиск останется обычным
    }
    setCustomerVehicles(result)
  }

  async function loadVehicleCustomers(vehicleId: string) {
    if (!vehicleId) {
      setCustomerPresets([])
      return
    }
    const result: SearchSelectItem<customersApi.Customer>[] = []
    const seen = new Set<string>()
    const push = (cid: string, label: string, sublabel?: string) => {
      if (!cid || seen.has(cid)) return
      seen.add(cid)
      result.push({ value: cid, label, sublabel, data: undefined })
    }
    try {
      const [woRes, coRes, dealRes] = await Promise.all([
        api.listWorkOrders({ vehicle_id: vehicleId, limit: 100 }),
        customerOrdersApi.listCustomerOrders({ limit: 200 }),
        dealsApi.listDeals({ limit: 200 }),
      ])
      for (const wo of woRes.work_orders || []) {
        push(wo.customer_id, wo.customer_name || wo.customer_id, wo.vehicle_label || wo.vehicle_vin)
      }
      for (const o of coRes.orders || []) {
        if (o.vehicle_id === vehicleId) push(o.customer_id, o.customer_name || o.customer_id, o.vehicle_label || o.vehicle_vin)
      }
      for (const d of dealRes.deals || []) {
        if (d.vehicle_id === vehicleId) push(d.customer_id, 'Клиент по сделке', d.vehicle_id)
      }
    } catch {
      // не критично — поиск останется обычным
    }
    setCustomerPresets(result)
  }

  const brandId = form.brand_id || ''
  const dealerPointId = form.dealer_point_id || ''

  useEffect(() => {
    if (!brandId || !dealerPointId) {
      setHourPrice('')
      setLaborRateNote('')
      return
    }
    brandsApi
      .resolveBrandLaborRate({
        brand_id: brandId,
        dealer_point_id: dealerPointId,
        repair_type: form.repair_type || 'commercial',
      })
      .then((r) => {
        if (r.found) {
          setHourPrice(r.hour_price)
          setLaborRateNote(`Гарантия: ${r.warranty_hour_price} ₽ · Коммерция: ${r.commercial_hour_price} ₽`)
        } else {
          setHourPrice('')
          setLaborRateNote('Тариф н/ч не задан для выбранного бренда и точки')
        }
      })
      .catch(() => {
        setHourPrice('')
        setLaborRateNote('')
      })
  }, [brandId, dealerPointId, form.repair_type])

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
            ? wo.labor.map((l) => ({
                ...l,
                work_id: l.work_id || '',
                executor_id: l.executor_id || '',
                description: l.description || l.work_name || '',
                display: l.work_code || l.work_name || l.description || '',
              }))
            : [emptyLabor()],
          parts: wo.parts
            ?.filter((p) => !p.issued)
            .map((p) => ({
              ...p,
              display: p.part_sku || p.part_name || p.description || '',
            })) || [],
        })
        setCustomerDisplay(wo.customer_name)
        setVehicleDisplay(wo.vehicle_label || wo.vehicle_vin)
        setOpenedAt(fromUnix(wo.opened_at))
        setClosedAt(fromUnix(wo.closed_at))
        if (wo.customer_id) void loadCustomerVehicles(wo.customer_id)
        if (wo.vehicle_id) void loadVehicleCustomers(wo.vehicle_id)
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

  function stripDisplay<T extends { display?: string }>(row: T): Omit<T, 'display'> {
    const { display, ...rest } = row
    void display
    return rest
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setSubmitting(true)
    setError(null)
    const payload: FormType = {
      ...form,
      opened_at: toUnix(openedAt),
      closed_at: toUnix(closedAt),
      labor: form.labor?.filter((l) => l.work_id).map(stripDisplay),
      parts: form.parts?.filter((p) => p.part_id).map(stripDisplay),
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

  function updateLabor(index: number, patch: Partial<LaborRow>) {
    setForm((f) => {
      const labor = [...(f.labor || [])]
      labor[index] = { ...labor[index], ...patch }
      return { ...f, labor }
    })
  }

  function updatePart(index: number, patch: Partial<PartRow>) {
    setForm((f) => {
      const partsRows = [...(f.parts || [])]
      partsRows[index] = { ...partsRows[index], ...patch }
      return { ...f, parts: partsRows }
    })
  }

  async function searchCustomers(query: string): Promise<SearchSelectItem<customersApi.Customer>[]> {
    const r = await customersApi.listCustomers({ limit: 10, search: query })
    return r.customers.map(customerItem)
  }

  async function searchVehicles(query: string): Promise<SearchSelectItem<vehiclesApi.Vehicle>[]> {
    const r = await vehiclesApi.listVehicles({ limit: 10, search: query })
    return r.vehicles.map(vehicleItem)
  }

  async function searchParts(query: string): Promise<SearchSelectItem<partsApi.Part>[]> {
    const r = await partsApi.listParts({ limit: 10, search: query })
    return r.parts.map(partItem)
  }

  async function searchWorks(query: string): Promise<SearchSelectItem<worksApi.Work>[]> {
    const r = await worksApi.listWorks({ limit: 10, search: query })
    return r.works.map(workItem)
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
            <SearchSelect<customersApi.Customer>
              value={form.customer_id || ''}
              displayValue={customerDisplay}
              placeholder="ФИО или email клиента…"
              clearable
              presetItems={customerPresets}
              presetHeader="Клиенты этого автомобиля"
              onChange={(v, item) => {
                setCustomerDisplay(item ? item.label : '')
                setForm((f) => ({ ...f, customer_id: v }))
                void loadCustomerVehicles(v)
              }}
              onSearch={searchCustomers}
            />
          </FormField>
          <FormField label="Автомобиль" required>
            <SearchSelect<vehiclesApi.Vehicle>
              value={form.vehicle_id || ''}
              displayValue={vehicleDisplay}
              placeholder="Марка, модель или VIN…"
              clearable
              presetItems={customerVehicles}
              presetHeader="Автомобили клиента"
              onChange={(v, item) => {
                setVehicleDisplay(item ? item.label : '')
                setForm((f) => ({
                  ...f,
                  vehicle_id: v,
                  brand_id: item?.data?.brand_id || undefined,
                  dealer_point_id: item?.data?.dealer_point_id || f.dealer_point_id,
                }))
                void loadVehicleCustomers(v)
              }}
              onSearch={searchVehicles}
            />
          </FormField>
          <FormField label="Дилерская точка">
            <NativeSelect
              value={form.dealer_point_id || ''}
              onChange={(e) => setForm({ ...form, dealer_point_id: e.target.value })}
            >
              <option value="">Не выбрана</option>
              {dealerPoints.map((p) => (
                <option key={p.id} value={p.id}>{p.name}</option>
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

        {laborRateNote && (
          <p className="text-sm text-muted-foreground">
            Стоимость н/ч: {hourPrice ? `${hourPrice} ₽` : '—'} ({laborRateNote})
          </p>
        )}

        <Card>
          <CardHeader className="flex flex-row items-center justify-between">
            <CardTitle className="text-base">Работы</CardTitle>
            <Button type="button" variant="outline" size="sm" onClick={() => setForm((f) => ({ ...f, labor: [...(f.labor || []), emptyLabor()] }))}>
              <Plus className="mr-1 h-4 w-4" /> Добавить
            </Button>
          </CardHeader>
          <CardContent className="space-y-3">
            {(form.labor || []).map((l, i) => (
              <div key={i} className="grid gap-2 rounded-md border p-3 md:grid-cols-6">
                <SearchSelect<worksApi.Work>
                  value={l.work_id || ''}
                  displayValue={l.display || ''}
                  placeholder="Код работы…"
                  clearable
                  onChange={(v, item) => {
                    updateLabor(i, {
                      work_id: v,
                      display: item ? item.label : '',
                      description: item?.data?.name || '',
                      quantity: item?.data?.labor_hours || '1',
                      unit_price: item ? (hourPrice || item.data?.unit_price || '0') : '0',
                    })
                  }}
                  onSearch={searchWorks}
                />
                <Input
                  placeholder="Название работы"
                  value={l.description || ''}
                  onChange={(e) => updateLabor(i, { description: e.target.value })}
                />
                <Input placeholder="Нормо-часы" value={l.quantity} onChange={(e) => updateLabor(i, { quantity: e.target.value })} />
                <Input placeholder="Цена н/ч" value={l.unit_price} onChange={(e) => updateLabor(i, { unit_price: e.target.value })} />
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
                <SearchSelect<partsApi.Part>
                  value={p.part_id || ''}
                  displayValue={p.display || ''}
                  placeholder="Артикул запчасти…"
                  clearable
                  onChange={(v, item) => {
                    updatePart(i, {
                      part_id: v,
                      display: item ? item.label : '',
                      description: item?.data?.name || '',
                      unit_price: item?.data?.price || '0',
                      warehouse_id: p.warehouse_id || item?.data?.warehouse_id || form.warehouse_id || '',
                    })
                  }}
                  onSearch={searchParts}
                />
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
