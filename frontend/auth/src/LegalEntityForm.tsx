import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import type { LegalEntityForm as FormType } from './dealerPointsApi'
import * as api from './dealerPointsApi'
import './Form.css'

export function LegalEntityForm() {
  const { id } = useParams()
  const navigate = useNavigate()
  const isNew = id === 'new' || !id
  const [loading, setLoading] = useState(!isNew)
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)
  const [form, setForm] = useState<FormType>({ name: '', inn: '', address: '' })

  useEffect(() => {
    if (isNew) return
    api.getLegalEntity(id!)
      .then((e) => setForm({ name: e.name, inn: e.inn || '', address: e.address || '' }))
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
    const payload = {
      name: form.name.trim(),
      inn: form.inn?.trim() || undefined,
      address: form.address?.trim() || undefined,
    }
    if (isNew) {
      api.createLegalEntity(payload)
        .then(() => navigate('/legal-entities', { replace: true }))
        .catch((err) => { setError(err.message); setSubmitting(false) })
    } else {
      api.updateLegalEntity(id!, payload)
        .then(() => navigate('/legal-entities', { replace: true }))
        .catch((err) => { setError(err.message); setSubmitting(false) })
    }
  }

  if (loading) return <div className="main loading">Загрузка…</div>

  return (
    <div className="form-card">
      <h1 className="form-title">{isNew ? 'Новое юридическое лицо' : 'Редактирование юр. лица'}</h1>
      <form onSubmit={handleSubmit} className="form">
        {error && <div className="form-error">{error}</div>}
        <label className="form-label">
          Название *
          <input
            value={form.name}
            onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))}
            required
            className="form-input"
            placeholder="ООО «Компания»"
          />
        </label>
        <label className="form-label">
          ИНН
          <input
            value={form.inn ?? ''}
            onChange={(e) => setForm((f) => ({ ...f, inn: e.target.value }))}
            className="form-input"
            placeholder="7707123456"
          />
        </label>
        <label className="form-label">
          Адрес
          <input
            value={form.address ?? ''}
            onChange={(e) => setForm((f) => ({ ...f, address: e.target.value }))}
            className="form-input"
          />
        </label>
        <div className="form-actions">
          <button type="submit" disabled={submitting} className="form-submit">
            {submitting ? 'Сохранение…' : (isNew ? 'Создать' : 'Сохранить')}
          </button>
          <button type="button" onClick={() => navigate('/legal-entities')} className="form-cancel">
            Отмена
          </button>
        </div>
      </form>
    </div>
  )
}
