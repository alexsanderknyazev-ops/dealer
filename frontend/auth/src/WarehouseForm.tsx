import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import type { WarehouseForm as FormType } from './dealerPointsApi'
import * as api from './dealerPointsApi'
import './Form.css'

export function WarehouseForm() {
  const { id } = useParams()
  const navigate = useNavigate()
  const isNew = id === 'new' || !id
  const [loading, setLoading] = useState(!isNew)
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)
  const [points, setPoints] = useState<api.DealerPoint[]>([])
  const [legalEntities, setLegalEntities] = useState<api.LegalEntity[]>([])
  const [form, setForm] = useState<FormType>({
    dealer_point_id: '',
    legal_entity_id: '',
    type: 'parts',
    name: '',
  })

  useEffect(() => {
    api.listDealerPoints({ limit: 200 }).then((r) => setPoints(r.dealer_points)).catch(() => setPoints([]))
  }, [])

  useEffect(() => {
    if (!form.dealer_point_id) {
      setLegalEntities([])
      return
    }
    api
      .listLegalEntitiesByDealerPoint(form.dealer_point_id)
      .then(setLegalEntities)
      .catch(() => setLegalEntities([]))
  }, [form.dealer_point_id])

  useEffect(() => {
    if (isNew) return
    api
      .getWarehouse(id!)
      .then((w) =>
        setForm({
          dealer_point_id: w.dealer_point_id,
          legal_entity_id: w.legal_entity_id,
          type: w.type,
          name: w.name,
        })
      )
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false))
  }, [id, isNew])

  useEffect(() => {
    if (!isNew || !form.dealer_point_id) return
    api
      .listLegalEntitiesByDealerPoint(form.dealer_point_id)
      .then(setLegalEntities)
      .catch(() => setLegalEntities([]))
  }, [isNew, form.dealer_point_id])

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!form.dealer_point_id || !form.legal_entity_id || !form.name?.trim()) {
      setError('Укажите дилерскую точку, юр. лицо и название склада')
      return
    }
    setError(null)
    setSubmitting(true)
    const payload = {
      dealer_point_id: form.dealer_point_id,
      legal_entity_id: form.legal_entity_id,
      type: form.type,
      name: form.name.trim(),
    }
    if (isNew) {
      api
        .createWarehouse(payload)
        .then(() => navigate('/warehouses', { replace: true }))
        .catch((err) => {
          setError(err.message)
          setSubmitting(false)
        })
    } else {
      api
        .updateWarehouse(id!, { name: form.name.trim() })
        .then(() => navigate('/warehouses', { replace: true }))
        .catch((err) => {
          setError(err.message)
          setSubmitting(false)
        })
    }
  }

  if (loading) return <div className="main loading">Загрузка…</div>

  return (
    <div className="form-card">
      <h1 className="form-title">{isNew ? 'Новый склад' : 'Редактирование склада'}</h1>
      <form onSubmit={handleSubmit} className="form">
        {error && <div className="form-error">{error}</div>}
        <label className="form-label">
          Дилерская точка *
          <select
            value={form.dealer_point_id}
            onChange={(e) =>
              setForm((f) => ({
                ...f,
                dealer_point_id: e.target.value,
                legal_entity_id: '',
              }))
            }
            required
            className="form-input"
            disabled={!isNew}
          >
            <option value="">— выберите —</option>
            {points.map((p) => (
              <option key={p.id} value={p.id}>
                {p.name}
              </option>
            ))}
          </select>
        </label>
        <label className="form-label">
          Юридическое лицо *
          <select
            value={form.legal_entity_id}
            onChange={(e) => setForm((f) => ({ ...f, legal_entity_id: e.target.value }))}
            required
            className="form-input"
            disabled={!isNew}
          >
            <option value="">— выберите —</option>
            {legalEntities.map((e) => (
              <option key={e.id} value={e.id}>
                {e.name} {e.inn ? `(ИНН ${e.inn})` : ''}
              </option>
            ))}
          </select>
        </label>
        <label className="form-label">
          Тип склада *
          <select
            value={form.type}
            onChange={(e) => setForm((f) => ({ ...f, type: e.target.value as 'cars' | 'parts' }))}
            className="form-input"
            disabled={!isNew}
          >
            <option value="cars">Автомобили</option>
            <option value="parts">Запчасти</option>
          </select>
        </label>
        <label className="form-label">
          Название склада *
          <input
            value={form.name}
            onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))}
            required
            className="form-input"
            placeholder="Например: Склад автомобилей — Москва"
          />
        </label>
        <div className="form-actions">
          <button type="submit" disabled={submitting} className="form-submit">
            {submitting ? 'Сохранение…' : isNew ? 'Создать' : 'Сохранить'}
          </button>
          <button type="button" onClick={() => navigate('/warehouses')} className="form-cancel">
            Отмена
          </button>
        </div>
      </form>
    </div>
  )
}
