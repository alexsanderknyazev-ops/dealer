import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import type { DealForm as DealFormType } from './dealsApi'
import * as api from './dealsApi'
import * as customersApi from './customersApi'
import * as vehiclesApi from './vehiclesApi'
import { useAuth } from './auth'
import { FormPage } from '@/components/common/FormPage'
import { FormField } from '@/components/common/FormField'
import { FormActions } from '@/components/common/FormActions'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Input } from '@/components/ui/input'
import { NativeSelect } from '@/components/ui/native-select'
import { Textarea } from '@/components/ui/textarea'

const DEAL_STAGES = ['draft', 'in_progress', 'paid', 'completed', 'cancelled'] as const

export function DealForm() {
  const { id } = useParams()
  const navigate = useNavigate()
  const { logout } = useAuth()
  const isNew = id === 'new' || !id
  const [loading, setLoading] = useState(!isNew)
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)
  const [fieldErrors, setFieldErrors] = useState<
    Partial<Record<'customer_id' | 'vehicle_id' | 'amount' | 'stage', string>>
  >({})
  const [customers, setCustomers] = useState<customersApi.Customer[]>([])
  const [vehicles, setVehicles] = useState<vehiclesApi.Vehicle[]>([])
  const [form, setForm] = useState<DealFormType>({
    customer_id: '',
    vehicle_id: '',
    amount: '',
    stage: 'draft',
    assigned_to: '',
    notes: '',
  })

  useEffect(() => {
    customersApi.listCustomers({ limit: 500 }).then((r) => setCustomers(r.customers)).catch(() => {})
    vehiclesApi.listVehicles({ limit: 500, status: 'available' }).then((r) => setVehicles(r.vehicles)).catch(() => {})
  }, [])

  useEffect(() => {
    if (isNew) return
    api
      .getDeal(id!)
      .then((d) => {
        setForm({
          customer_id: d.customer_id,
          vehicle_id: d.vehicle_id,
          amount: d.amount || '',
          stage: d.stage || 'draft',
          assigned_to: d.assigned_to || '',
          notes: d.notes || '',
        })
      })
      .catch(async (e) => {
        if (e instanceof api.ApiError && (e.status === 401 || e.status === 403)) {
          await logout()
          navigate('/login', { replace: true })
          return
        }
        setError(e instanceof Error ? e.message : 'Ошибка загрузки сделки')
      })
      .finally(() => setLoading(false))
  }, [id, isNew, logout, navigate])

  function validateDealForm(nextForm: DealFormType) {
    const nextErrors: Partial<Record<'customer_id' | 'vehicle_id' | 'amount' | 'stage', string>> = {}
    if (!nextForm.customer_id) nextErrors.customer_id = 'Выберите клиента'
    if (!nextForm.vehicle_id) nextErrors.vehicle_id = 'Выберите автомобиль'
    if (nextForm.amount && !/^\d+(\.\d{1,2})?$/.test(nextForm.amount.trim())) {
      nextErrors.amount = 'Укажите сумму в формате 10000 или 10000.50'
    }
    if (nextForm.amount && Number(nextForm.amount) < 0) {
      nextErrors.amount = 'Сумма не может быть отрицательной'
    }
    if (nextForm.stage && !DEAL_STAGES.includes(nextForm.stage as (typeof DEAL_STAGES)[number])) {
      nextErrors.stage = 'Выберите корректный этап'
    }
    return nextErrors
  }

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    const nextFieldErrors = validateDealForm(form)
    setFieldErrors(nextFieldErrors)
    if (Object.keys(nextFieldErrors).length > 0) {
      setError('Исправьте ошибки в форме перед сохранением')
      return
    }
    setError(null)
    setSubmitting(true)
    const payload = {
      customer_id: form.customer_id,
      vehicle_id: form.vehicle_id,
      amount: (form.amount ?? '').trim() !== '' ? (form.amount ?? '').trim() : '0',
      stage: form.stage || undefined,
      assigned_to: form.assigned_to || undefined,
      notes: form.notes || undefined,
    }
    const save = isNew ? api.createDeal(payload) : api.updateDeal(id!, payload)
    save
      .then(() => navigate('/deals', { replace: true }))
      .catch(async (err) => {
        if (err instanceof api.ApiError && (err.status === 401 || err.status === 403)) {
          await logout()
          navigate('/login', { replace: true })
          return
        }
        setError(err instanceof Error ? err.message : 'Ошибка сохранения сделки')
        setSubmitting(false)
      })
  }

  const isFormValid = Object.keys(validateDealForm(form)).length === 0

  return (
    <FormPage title={isNew ? 'Новая сделка' : 'Редактирование сделки'} loading={loading}>
      <form onSubmit={handleSubmit} className="space-y-4">
        {error && (
          <Alert variant="destructive">
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}
        <FormField label="Клиент" htmlFor="customer_id" required error={fieldErrors.customer_id}>
          <NativeSelect
            id="customer_id"
            value={form.customer_id}
            aria-invalid={!!fieldErrors.customer_id}
            onChange={(e) => {
              const value = e.target.value
              setForm((f) => ({ ...f, customer_id: value }))
              setFieldErrors((prev) => ({ ...prev, customer_id: value ? undefined : 'Выберите клиента' }))
            }}
            required
          >
            <option value="">— Выберите —</option>
            {customers.map((c) => (
              <option key={c.id} value={c.id}>
                {c.name} ({c.email || c.phone || c.id.slice(0, 8)})
              </option>
            ))}
          </NativeSelect>
        </FormField>
        <FormField label="Автомобиль" htmlFor="vehicle_id" required error={fieldErrors.vehicle_id}>
          <NativeSelect
            id="vehicle_id"
            value={form.vehicle_id}
            aria-invalid={!!fieldErrors.vehicle_id}
            onChange={(e) => {
              const value = e.target.value
              setForm((f) => ({ ...f, vehicle_id: value }))
              setFieldErrors((prev) => ({ ...prev, vehicle_id: value ? undefined : 'Выберите автомобиль' }))
            }}
            required
          >
            <option value="">— Выберите —</option>
            {vehicles.map((v) => (
              <option key={v.id} value={v.id}>
                {v.make} {v.model} ({v.year}) — {v.vin}
              </option>
            ))}
          </NativeSelect>
        </FormField>
        <FormField label="Сумма" htmlFor="amount" error={fieldErrors.amount}>
          <Input
            id="amount"
            type="text"
            value={form.amount}
            placeholder="0"
            aria-invalid={!!fieldErrors.amount}
            onChange={(e) => {
              const value = e.target.value
              setForm((f) => ({ ...f, amount: value }))
              const amountErrors = validateDealForm({ ...form, amount: value })
              setFieldErrors((prev) => ({ ...prev, amount: amountErrors.amount }))
            }}
          />
        </FormField>
        <FormField label="Этап" htmlFor="stage" error={fieldErrors.stage}>
          <NativeSelect
            id="stage"
            value={form.stage}
            aria-invalid={!!fieldErrors.stage}
            onChange={(e) => {
              const value = e.target.value
              setForm((f) => ({ ...f, stage: value }))
              const stageErrors = validateDealForm({ ...form, stage: value })
              setFieldErrors((prev) => ({ ...prev, stage: stageErrors.stage }))
            }}
          >
            <option value="draft">Черновик</option>
            <option value="in_progress">В работе</option>
            <option value="paid">Оплачено</option>
            <option value="completed">Завершено</option>
            <option value="cancelled">Отменено</option>
          </NativeSelect>
        </FormField>
        <FormField label="Заметки" htmlFor="notes">
          <Textarea
            id="notes"
            value={form.notes}
            onChange={(e) => setForm((f) => ({ ...f, notes: e.target.value }))}
            rows={3}
          />
        </FormField>
        <FormActions
          submitting={submitting}
          submitLabel={isNew ? 'Создать' : 'Сохранить'}
          onCancel={() => navigate('/deals')}
          disabled={!isFormValid}
        />
      </form>
    </FormPage>
  )
}
