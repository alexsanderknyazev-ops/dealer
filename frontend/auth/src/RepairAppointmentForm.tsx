import { useEffect, useState } from 'react'
import { useNavigate, useParams, useSearchParams } from 'react-router-dom'
import { Plus, Trash2 } from 'lucide-react'
import * as api from './repairAppointmentsApi'
import * as customersApi from './customersApi'
import * as vehiclesApi from './vehiclesApi'
import * as partsApi from './partsApi'
import * as worksApi from './worksApi'
import * as dealerPointsApi from './dealerPointsApi'
import * as workOrdersApi from './workOrdersApi'
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

type LaborLine = api.LaborLineInput & { display?: string }
type PartLine = api.PartLineInput & { display?: string }

type AppointmentFormState = Omit<api.RepairAppointmentForm, 'labor' | 'parts'> & {
  labor?: LaborLine[]
  parts?: PartLine[]
}

const emptyLabor = (): LaborLine => ({ work_id: '', description: '', quantity: '1', unit_price: '0' })
const emptyPart = (): PartLine => ({ part_id: '', warehouse_id: '', quantity: '1', unit_price: '0', notes: '' })

function fromUnix(ts?: number): string {
  if (!ts) return ''
  const d = new Date(ts * 1000)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`
}

function toUnix(dt: string): number {
  return Math.floor(Date.parse(dt) / 1000)
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

function stripDisplay<T extends { display?: string }>(row: T): Omit<T, 'display'> {
  const { display, ...rest } = row
  void display
  return rest
}

export function RepairAppointmentForm() {
  const { id } = useParams()
  const [searchParams] = useSearchParams()
  const navigate = useNavigate()
  const { logout, user } = useAuth()
  const isNew = id === 'new' || !id
  const [loading, setLoading] = useState(!isNew)
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)
  const [warehouses, setWarehouses] = useState<dealerPointsApi.Warehouse[]>([])
  const [customerDisplay, setCustomerDisplay] = useState('')
  const [vehicleDisplay, setVehicleDisplay] = useState('')
  const [customerVehicles, setCustomerVehicles] = useState<SearchSelectItem<vehiclesApi.Vehicle>[]>([])
  const [customerPresets, setCustomerPresets] = useState<SearchSelectItem<customersApi.Customer>[]>([])

  const startParam = searchParams.get('start')
  const endParam = searchParams.get('end')
  const defaultStart = startParam ? Number(startParam) : undefined
  const defaultEnd = endParam ? Number(endParam) : undefined

  const [startAt, setStartAt] = useState(fromUnix(defaultStart))
  const [endAt, setEndAt] = useState(fromUnix(defaultEnd))
  const [form, setForm] = useState<AppointmentFormState>({
    customer_id: '',
    vehicle_id: '',
    warehouse_id: '',
    scheduled_start: defaultStart || 0,
    scheduled_end: defaultEnd || 0,
    complaint: '',
    notes: '',
    labor: [emptyLabor()],
    parts: [],
  })

  useEffect(() => {
    dealerPointsApi.listWarehouses({ limit: 500 }).then((r) => setWarehouses(r.warehouses)).catch(() => {})
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
        workOrdersApi.listWorkOrders({ customer_id: customerId, limit: 100 }),
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
        workOrdersApi.listWorkOrders({ vehicle_id: vehicleId, limit: 100 }),
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

  useEffect(() => {
    if (isNew) return
    api
      .getRepairAppointment(id!)
      .then((a) => {
        if (a.work_order_id || a.status === 'cancelled' || a.status === 'completed') {
          setError('Запись нельзя редактировать')
          return
        }
        setStartAt(fromUnix(a.scheduled_start))
        setEndAt(fromUnix(a.scheduled_end))
        setCustomerDisplay(a.customer_name)
        setVehicleDisplay(a.vehicle_label || a.vehicle_vin)
        setForm({
          customer_id: a.customer_id,
          vehicle_id: a.vehicle_id,
          warehouse_id: a.warehouse_id || '',
          scheduled_start: a.scheduled_start,
          scheduled_end: a.scheduled_end,
          complaint: a.complaint || '',
          notes: a.notes || '',
          labor: a.labor?.length
            ? a.labor.map((l) => ({
                work_id: l.work_id,
                description: l.description,
                quantity: l.quantity,
                unit_price: l.unit_price,
                sort_order: l.sort_order,
                display: l.work_code || l.work_name || l.description || '',
              }))
            : [emptyLabor()],
          parts: a.parts?.map((p) => ({
            part_id: p.part_id,
            warehouse_id: p.warehouse_id,
            quantity: p.quantity,
            unit_price: p.unit_price,
            notes: p.notes,
            sort_order: p.sort_order,
            display: p.part_sku || p.part_name || '',
          })) || [],
        })
        if (a.customer_id) void loadCustomerVehicles(a.customer_id)
        if (a.vehicle_id) void loadVehicleCustomers(a.vehicle_id)
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
    if (!form.customer_id || !form.vehicle_id) {
      setError('Укажите клиента и автомобиль')
      return
    }
    if (!startAt || !endAt) {
      setError('Укажите время начала и окончания')
      return
    }
    const scheduled_start = toUnix(startAt)
    const scheduled_end = toUnix(endAt)
    if (scheduled_end <= scheduled_start) {
      setError('Время окончания должно быть позже начала')
      return
    }
    setSubmitting(true)
    setError(null)
    const payload: api.RepairAppointmentForm = {
      ...form,
      scheduled_start,
      scheduled_end,
      labor: (form.labor || []).filter((l) => l.work_id || l.description?.trim()).map(stripDisplay),
      parts: (form.parts || []).filter((p) => p.part_id && p.warehouse_id).map(stripDisplay),
    }
    try {
      const saved = isNew
        ? await api.createRepairAppointment(payload, user?.userId)
        : await api.updateRepairAppointment(id!, payload)
      navigate(`/repair-appointments/${saved.id}`)
    } catch (err) {
      if (err instanceof api.ApiError && (err.status === 401 || err.status === 403)) {
        await logout()
        navigate('/login', { replace: true })
        return
      }
      setError(err instanceof Error ? err.message : 'Ошибка сохранения')
    } finally {
      setSubmitting(false)
    }
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
    <FormPage title={isNew ? 'Новая запись на ремонт' : 'Редактирование записи'} loading={loading}>
      {error && (
        <Alert variant="destructive" className="mb-4">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}
      <form onSubmit={handleSubmit} className="space-y-6">
        <Card>
          <CardHeader><CardTitle className="text-base">Клиент и время</CardTitle></CardHeader>
          <CardContent className="grid gap-4 sm:grid-cols-2">
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
                  setForm((f) => ({ ...f, customer_id: v, vehicle_id: f.vehicle_id }))
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
                  setForm((f) => ({ ...f, vehicle_id: v }))
                  void loadVehicleCustomers(v)
                }}
                onSearch={searchVehicles}
              />
            </FormField>
            <FormField label="Склад (для запчастей)">
              <NativeSelect
                value={form.warehouse_id || ''}
                onChange={(e) => setForm((f) => ({ ...f, warehouse_id: e.target.value }))}
              >
                <option value="">По умолчанию</option>
                {warehouses.map((w) => (
                  <option key={w.id} value={w.id}>{w.name}</option>
                ))}
              </NativeSelect>
            </FormField>
            <div />
            <FormField label="Начало" required>
              <Input type="datetime-local" value={startAt} onChange={(e) => setStartAt(e.target.value)} />
            </FormField>
            <FormField label="Окончание" required>
              <Input type="datetime-local" value={endAt} onChange={(e) => setEndAt(e.target.value)} />
            </FormField>
            <FormField label="Жалоба клиента" className="sm:col-span-2">
              <Textarea value={form.complaint || ''} onChange={(e) => setForm((f) => ({ ...f, complaint: e.target.value }))} />
            </FormField>
            <FormField label="Примечания" className="sm:col-span-2">
              <Textarea value={form.notes || ''} onChange={(e) => setForm((f) => ({ ...f, notes: e.target.value }))} />
            </FormField>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between">
            <CardTitle className="text-base">Работы</CardTitle>
            <Button type="button" variant="outline" size="sm" onClick={() => setForm((f) => ({ ...f, labor: [...(f.labor || []), emptyLabor()] }))}>
              <Plus className="mr-1 h-4 w-4" /> Добавить
            </Button>
          </CardHeader>
          <CardContent className="space-y-3">
            {(form.labor || []).map((line, i) => (
              <div key={i} className="grid gap-2 rounded border p-3 sm:grid-cols-5">
                <SearchSelect<worksApi.Work>
                  value={line.work_id || ''}
                  displayValue={line.display || ''}
                  placeholder="Код работы…"
                  clearable
                  onChange={(v, item) => {
                    setForm((f) => {
                      const labor = [...(f.labor || [])]
                      labor[i] = {
                        ...labor[i],
                        work_id: v,
                        display: item ? item.label : '',
                        description: item?.data?.name || labor[i].description,
                        quantity: item?.data?.labor_hours || labor[i].quantity,
                        unit_price: item?.data?.unit_price || labor[i].unit_price,
                      }
                      return { ...f, labor }
                    })
                  }}
                  onSearch={searchWorks}
                />
                <Input
                  placeholder="Описание"
                  value={line.description || ''}
                  onChange={(e) => {
                    const labor = [...(form.labor || [])]
                    labor[i] = { ...labor[i], description: e.target.value }
                    setForm((f) => ({ ...f, labor }))
                  }}
                />
                <Input
                  type="number"
                  step="0.01"
                  placeholder="Кол-во"
                  value={line.quantity || ''}
                  onChange={(e) => {
                    const labor = [...(form.labor || [])]
                    labor[i] = { ...labor[i], quantity: e.target.value }
                    setForm((f) => ({ ...f, labor }))
                  }}
                />
                <Input
                  type="number"
                  step="0.01"
                  placeholder="Цена"
                  value={line.unit_price || ''}
                  onChange={(e) => {
                    const labor = [...(form.labor || [])]
                    labor[i] = { ...labor[i], unit_price: e.target.value }
                    setForm((f) => ({ ...f, labor }))
                  }}
                />
                <Button type="button" variant="ghost" size="icon" onClick={() => setForm((f) => ({ ...f, labor: (f.labor || []).filter((_, j) => j !== i) }))}>
                  <Trash2 className="h-4 w-4" />
                </Button>
              </div>
            ))}
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between">
            <CardTitle className="text-base">Запчасти</CardTitle>
            <Button type="button" variant="outline" size="sm" onClick={() => setForm((f) => ({ ...f, parts: [...(f.parts || []), { ...emptyPart(), warehouse_id: f.warehouse_id || '' }] }))}>
              <Plus className="mr-1 h-4 w-4" /> Добавить
            </Button>
          </CardHeader>
          <CardContent className="space-y-3">
            {(form.parts || []).map((line, i) => (
              <div key={i} className="grid gap-2 rounded border p-3 sm:grid-cols-5">
                <SearchSelect<partsApi.Part>
                  value={line.part_id || ''}
                  displayValue={line.display || ''}
                  placeholder="Артикул запчасти…"
                  clearable
                  onChange={(v, item) => {
                    setForm((f) => {
                      const partsLines = [...(f.parts || [])]
                      partsLines[i] = {
                        ...partsLines[i],
                        part_id: v,
                        display: item ? item.label : '',
                        warehouse_id: partsLines[i].warehouse_id || item?.data?.warehouse_id || f.warehouse_id || '',
                        unit_price: item?.data?.price || partsLines[i].unit_price,
                      }
                      return { ...f, parts: partsLines }
                    })
                  }}
                  onSearch={searchParts}
                />
                <NativeSelect
                  value={line.warehouse_id}
                  onChange={(e) => {
                    const partsLines = [...(form.parts || [])]
                    partsLines[i] = { ...partsLines[i], warehouse_id: e.target.value }
                    setForm((f) => ({ ...f, parts: partsLines }))
                  }}
                >
                  <option value="">Склад</option>
                  {warehouses.map((w) => <option key={w.id} value={w.id}>{w.name}</option>)}
                </NativeSelect>
                <Input
                  type="number"
                  step="0.01"
                  placeholder="Кол-во"
                  value={line.quantity || ''}
                  onChange={(e) => {
                    const partsLines = [...(form.parts || [])]
                    partsLines[i] = { ...partsLines[i], quantity: e.target.value }
                    setForm((f) => ({ ...f, parts: partsLines }))
                  }}
                />
                <Input
                  type="number"
                  step="0.01"
                  placeholder="Цена"
                  value={line.unit_price || ''}
                  onChange={(e) => {
                    const partsLines = [...(form.parts || [])]
                    partsLines[i] = { ...partsLines[i], unit_price: e.target.value }
                    setForm((f) => ({ ...f, parts: partsLines }))
                  }}
                />
                <Button type="button" variant="ghost" size="icon" onClick={() => setForm((f) => ({ ...f, parts: (f.parts || []).filter((_, j) => j !== i) }))}>
                  <Trash2 className="h-4 w-4" />
                </Button>
              </div>
            ))}
            {(form.parts || []).length === 0 && <p className="text-sm text-muted-foreground">Запчасти не добавлены</p>}
          </CardContent>
        </Card>

        <FormActions submitting={submitting} submitLabel={isNew ? 'Создать' : 'Сохранить'} onCancel={() => navigate('/repair-appointments')} />
      </form>
    </FormPage>
  )
}
