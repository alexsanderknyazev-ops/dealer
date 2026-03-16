import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import type { DealerPointForm as FormType } from './dealerPointsApi'
import * as api from './dealerPointsApi'
import './Form.css'

export function DealerPointForm() {
  const { id } = useParams()
  const navigate = useNavigate()
  const isNew = id === 'new' || !id
  const [loading, setLoading] = useState(!isNew)
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)
  const [form, setForm] = useState<FormType>({ name: '', address: '' })

  useEffect(() => {
    if (isNew) return
    api.getDealerPoint(id!)
      .then((d) => setForm({ name: d.name, address: d.address || '' }))
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false))
  }, [id, isNew])

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!form.name?.trim()) {
      setError('Укажите название')
      return
    }
    setError(null)
    setSubmitting(true)
    const payload = { name: form.name.trim(), address: form.address || undefined }
    if (isNew) {
      api.createDealerPoint(payload)
        .then(() => navigate('/dealer-points', { replace: true }))
        .catch((err) => { setError(err.message); setSubmitting(false) })
    } else {
      api.updateDealerPoint(id!, payload)
        .then(() => navigate('/dealer-points', { replace: true }))
        .catch((err) => { setError(err.message); setSubmitting(false) })
    }
  }

  if (loading) return <div className="main loading">Загрузка…</div>

  return (
    <div className="form-card">
      <h1 className="form-title">{isNew ? 'Новая дилерская точка' : 'Редактирование дилерской точки'}</h1>
      <form onSubmit={handleSubmit} className="form">
        {error && <div className="form-error">{error}</div>}
        <label className="form-label">
          Название *
          <input
            value={form.name}
            onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))}
            required
            className="form-input"
            placeholder="Например: ДЦ Москва Ленинградское ш."
          />
        </label>
        <label className="form-label">
          Адрес
          <input
            value={form.address ?? ''}
            onChange={(e) => setForm((f) => ({ ...f, address: e.target.value }))}
            className="form-input"
            placeholder="г. Москва, ул. Примерная, 1"
          />
        </label>
        <div className="form-actions">
          <button type="submit" disabled={submitting} className="form-submit">
            {submitting ? 'Сохранение…' : (isNew ? 'Создать' : 'Сохранить')}
          </button>
          <button type="button" onClick={() => navigate('/dealer-points')} className="form-cancel">
            Отмена
          </button>
        </div>
      </form>
    </div>
  )
}
