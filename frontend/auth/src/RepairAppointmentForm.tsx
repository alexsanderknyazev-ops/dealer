import { useEffect, useState } from 'react'
import { useNavigate, useParams, useSearchParams } from 'react-router-dom'
import { Plus, Trash2 } from 'lucide-react'
import * as api from './repairAppointmentsApi'
import * as customersApi from './customersApi'
import * as vehiclesApi from './vehiclesApi'
import * as partsApi from './partsApi'
import * as worksApi from './worksApi'
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

const emptyLabor = (): api.LaborLineInput => ({ work_id: '', description: '', quantity: '1', unit_price: '0' })
const emptyPart = (): api.PartLineInput => ({ part_id: '', warehouse_id: '', quantity: '1', unit_price: '0', notes: '' })

function fromUnix(ts?: number): string {
  if (!ts) return ''
  const d = new Date(ts * 1000)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`
}

function toUnix(dt: string): number {
  return Math.floor(Date.parse(dt) / 1000)
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
  const [customers, setCustomers] = useState<customersApi.Customer[]>([])
  const [vehicles, setVehicles] = useState<vehiclesApi.Vehicle[]>([])
  const [parts, setParts] = useState<partsApi.Part[]>([])
  const [works, setWorks] = useState<worksApi.Work[]>([])
  const [warehouses, setWarehouses] = useState<dealerPointsApi.Warehouse[]>([])

  const startParam = searchParams.get('start')
  const endParam = searchParams.get('end')
  const defaultStart = startParam ? Number(startParam) : undefined
  const defaultEnd = endParam ? Number(endParam) : undefined

  const [startAt, setStartAt] = useState(fromUnix(defaultStart))
  const [endAt, setEndAt] = useState(fromUnix(defaultEnd))
  const [form, setForm] = useState<api.RepairAppointmentForm>({
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
    customersApi.listCustomers({ limit: 500 }).then((r) => setCustomers(r.customers)).catch(() => {})
    vehiclesApi.listVehicles({ limit: 500 }).then((r) => setVehicles(r.vehicles)).catch(() => {})
    partsApi.listParts({ limit: 500 }).then((r) => setParts(r.parts)).catch(() => {})
    worksApi.listWorks({ limit: 500 }).then((r) => setWorks(r.works)).catch(() => {})
    dealerPointsApi.listWarehouses({ limit: 500 }).then((r) => setWarehouses(r.warehouses)).catch(() => {})
  }, [])

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
              }))
            : [emptyLabor()],
          parts: a.parts?.map((p) => ({
            part_id: p.part_id,
            warehouse_id: p.warehouse_id,
            quantity: p.quantity,
            unit_price: p.unit_price,
            notes: p.notes,
            sort_order: p.sort_order,
          })) || [],
        })
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
      labor: (form.labor || []).filter((l) => l.work_id || l.description?.trim()),
      parts: (form.parts || []).filter((p) => p.part_id && p.warehouse_id),
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
              <NativeSelect
                value={form.customer_id}
                onChange={(e) => setForm((f) => ({ ...f, customer_id: e.target.value, vehicle_id: '' }))}
              >
                <option value="">Выберите клиента</option>
                {customers.map((c) => (
                  <option key={c.id} value={c.id}>{c.name || c.email || c.id}</option>
                ))}
              </NativeSelect>
            </FormField>
            <FormField label="Автомобиль" required>
              <NativeSelect
                value={form.vehicle_id}
                onChange={(e) => setForm((f) => ({ ...f, vehicle_id: e.target.value }))}
              >
                <option value="">Выберите автомобиль</option>
                {vehicles.map((v) => (
                  <option key={v.id} value={v.id}>
                    {v.make} {v.model} {v.vin ? `(${v.vin})` : ''}
                  </option>
                ))}
              </NativeSelect>
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
                <NativeSelect
                  value={line.work_id || ''}
                  onChange={(e) => {
                    const work = works.find((w) => w.id === e.target.value)
                    setForm((f) => {
                      const labor = [...(f.labor || [])]
                      labor[i] = { ...labor[i], work_id: e.target.value, description: work?.name || labor[i].description }
                      return { ...f, labor }
                    })
                  }}
                >
                  <option value="">Работа</option>
                  {works.map((w) => <option key={w.id} value={w.id}>{w.code} — {w.name}</option>)}
                </NativeSelect>
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
                <NativeSelect
                  value={line.part_id}
                  onChange={(e) => {
                    const partsLines = [...(form.parts || [])]
                    partsLines[i] = { ...partsLines[i], part_id: e.target.value }
                    setForm((f) => ({ ...f, parts: partsLines }))
                  }}
                >
                  <option value="">Запчасть</option>
                  {parts.map((p) => <option key={p.id} value={p.id}>{p.sku} — {p.name}</option>)}
                </NativeSelect>
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
