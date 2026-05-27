import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import type { CustomerForm as CustomerFormType } from './customersApi'
import * as api from './customersApi'
import './Form.css'

export function CustomerForm() {
  const { id } = useParams()
  const navigate = useNavigate()
  const isNew = id === 'new' || !id
  const [loading, setLoading] = useState(!isNew)
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)
  const [form, setForm] = useState<CustomerFormType>({
    name: '',
    email: '',
    phone: '',
    customer_type: 'individual',
    inn: '',
    address: '',
    notes: '',
  })

  useEffect(() => {
    if (isNew) return
    api.getCustomer(id!)
      .then((c) => {
        setForm({
          name: c.name,
          email: c.email || '',
          phone: c.phone || '',
          customer_type: c.customer_type || 'individual',
          inn: c.inn || '',
          address: c.address || '',
          notes: c.notes || '',
        })
      })
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false))
  }, [id, isNew])

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError(null)
    setSubmitting(true)
    const payload = {
      name: form.name,
      email: form.email || undefined,
      phone: form.phone || undefined,
      customer_type: form.customer_type || undefined,
      inn: form.inn || undefined,
      address: form.address || undefined,
      notes: form.notes || undefined,
    }
    if (isNew) {
      api.createCustomer(payload)
        .then(() => navigate('/customers', { replace: true }))
        .catch((err) => { setError(err.message); setSubmitting(false) })
    } else {
      api.updateCustomer(id!, payload)
        .then(() => navigate('/customers', { replace: true }))
        .catch((err) => { setError(err.message); setSubmitting(false) })
    }
  }

  if (loading) return <div className="main loading">Загрузка…</div>

  return (
    <div className="form-card">
      <h1 className="form-title">{isNew ? 'Новый клиент' : 'Редактирование клиента'}</h1>
      <form onSubmit={handleSubmit} className="form">
        {error && <div className="form-error">{error}</div>}
        <label className="form-label">
          Имя *
          <input
            value={form.name}
            onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))}
            required
            className="form-input"
          />
        </label>
        <label className="form-label">
          Email
          <input
            type="email"
            value={form.email}
            onChange={(e) => setForm((f) => ({ ...f, email: e.target.value }))}
            className="form-input"
          />
        </label>
        <label className="form-label">
          Телефон
          <input
            type="tel"
            value={form.phone}
            onChange={(e) => setForm((f) => ({ ...f, phone: e.target.value }))}
            className="form-input"
          />
        </label>
        <label className="form-label">
          Тип
          <select
            value={form.customer_type}
            onChange={(e) => setForm((f) => ({ ...f, customer_type: e.target.value }))}
            className="form-input"
          >
            <option value="individual">Физ. лицо</option>
            <option value="legal">Юр. лицо</option>
          </select>
        </label>
        <label className="form-label">
          ИНН
          <input
            value={form.inn}
            onChange={(e) => setForm((f) => ({ ...f, inn: e.target.value }))}
            className="form-input"
          />
        </label>
        <label className="form-label">
          Адрес
          <input
            value={form.address}
            onChange={(e) => setForm((f) => ({ ...f, address: e.target.value }))}
            className="form-input"
          />
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
          <button type="button" onClick={() => navigate('/customers')} className="form-cancel">
            Отмена
          </button>
        </div>
      </form>
    </div>
  )
}
