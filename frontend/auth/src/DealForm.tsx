import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import type { DealForm as DealFormType } from './dealsApi'
import * as api from './dealsApi'
import * as customersApi from './customersApi'
import * as vehiclesApi from './vehiclesApi'
import './Form.css'

export function DealForm() {
  const { id } = useParams()
  const navigate = useNavigate()
  const isNew = id === 'new' || !id
  const [loading, setLoading] = useState(!isNew)
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)
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
    api.getDeal(id!)
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
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false))
  }, [id, isNew])

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!form.customer_id || !form.vehicle_id) {
      setError('Выберите клиента и автомобиль')
      return
    }
    setError(null)
    setSubmitting(true)
    const payload = {
      customer_id: form.customer_id,
      vehicle_id: form.vehicle_id,
      amount: form.amount || undefined,
      stage: form.stage || undefined,
      assigned_to: form.assigned_to || undefined,
      notes: form.notes || undefined,
    }
    if (isNew) {
      api.createDeal(payload)
        .then(() => navigate('/deals', { replace: true }))
        .catch((err) => { setError(err.message); setSubmitting(false) })
    } else {
      api.updateDeal(id!, payload)
        .then(() => navigate('/deals', { replace: true }))
        .catch((err) => { setError(err.message); setSubmitting(false) })
    }
  }

  if (loading) return <div className="main loading">Загрузка…</div>

  return (
    <div className="form-card">
      <h1 className="form-title">{isNew ? 'Новая сделка' : 'Редактирование сделки'}</h1>
      <form onSubmit={handleSubmit} className="form">
        {error && <div className="form-error">{error}</div>}
        <label className="form-label">
          Клиент *
          <select
            value={form.customer_id}
            onChange={(e) => setForm((f) => ({ ...f, customer_id: e.target.value }))}
            required
            className="form-input"
          >
            <option value="">— Выберите —</option>
            {customers.map((c) => (
              <option key={c.id} value={c.id}>{c.name} ({c.email || c.phone || c.id.slice(0, 8)})</option>
            ))}
          </select>
        </label>
        <label className="form-label">
          Автомобиль *
          <select
            value={form.vehicle_id}
            onChange={(e) => setForm((f) => ({ ...f, vehicle_id: e.target.value }))}
            required
            className="form-input"
          >
            <option value="">— Выберите —</option>
            {vehicles.map((v) => (
              <option key={v.id} value={v.id}>{v.make} {v.model} ({v.year}) — {v.vin}</option>
            ))}
          </select>
        </label>
        <label className="form-label">
          Сумма
          <input
            type="text"
            value={form.amount}
            onChange={(e) => setForm((f) => ({ ...f, amount: e.target.value }))}
            placeholder="0"
            className="form-input"
          />
        </label>
        <label className="form-label">
          Этап
          <select
            value={form.stage}
            onChange={(e) => setForm((f) => ({ ...f, stage: e.target.value }))}
            className="form-input"
          >
            <option value="draft">Черновик</option>
            <option value="in_progress">В работе</option>
            <option value="paid">Оплачено</option>
            <option value="completed">Завершено</option>
            <option value="cancelled">Отменено</option>
          </select>
        </label>
        <label className="form-label">
          Заметки
          <textarea
            value={form.notes}
            onChange={(e) => setForm((f) => ({ ...f, notes: e.target.value }))}
            className="form-input"
            rows={3}
          />
        </label>
        <div className="form-actions">
          <button type="submit" disabled={submitting} className="form-submit">
            {submitting ? 'Сохранение…' : (isNew ? 'Создать' : 'Сохранить')}
          </button>
          <button type="button" onClick={() => navigate('/deals')} className="form-cancel">
            Отмена
          </button>
        </div>
      </form>
    </div>
  )
}
