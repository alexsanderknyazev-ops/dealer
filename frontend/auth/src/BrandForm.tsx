import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import type { BrandForm as BrandFormType } from './brandsApi'
import * as api from './brandsApi'
import './Form.css'

export function BrandForm() {
  const { id } = useParams()
  const navigate = useNavigate()
  const isNew = id === 'new' || !id
  const [loading, setLoading] = useState(!isNew)
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)
  const [form, setForm] = useState<BrandFormType>({ name: '' })

  useEffect(() => {
    if (isNew) return
    api.getBrand(id!)
      .then((b) => setForm({ name: b.name }))
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
    const payload = { name: form.name.trim() }
    if (isNew) {
      api.createBrand(payload)
        .then(() => navigate('/brands', { replace: true }))
        .catch((err) => { setError(err.message); setSubmitting(false) })
    } else {
      api.updateBrand(id!, payload)
        .then(() => navigate('/brands', { replace: true }))
        .catch((err) => { setError(err.message); setSubmitting(false) })
    }
  }

  if (loading) return <div className="main loading">Загрузка…</div>

  return (
    <div className="form-card">
      <h1 className="form-title">{isNew ? 'Новый бренд' : 'Редактирование бренда'}</h1>
      <form onSubmit={handleSubmit} className="form">
        {error && <div className="form-error">{error}</div>}
        <label className="form-label">
          Название *
          <input
            value={form.name}
            onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))}
            required
            placeholder="Например: BMW"
            className="form-input"
          />
        </label>
        <div className="form-actions">
          <button type="submit" disabled={submitting} className="form-submit">
            {submitting ? 'Сохранение…' : (isNew ? 'Создать' : 'Сохранить')}
          </button>
          <button type="button" onClick={() => navigate('/brands')} className="form-cancel">
            Отмена
          </button>
        </div>
      </form>
    </div>
  )
}
