import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import type { VehicleForm as VehicleFormType } from './vehiclesApi'
import * as api from './vehiclesApi'
import * as brandsApi from './brandsApi'
import * as dealerPointsApi from './dealerPointsApi'
import './Form.css'

export function VehicleForm() {
  const { id } = useParams()
  const navigate = useNavigate()
  const isNew = id === 'new' || !id
  const [loading, setLoading] = useState(!isNew)
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)
  const [brands, setBrands] = useState<brandsApi.Brand[]>([])
  const [points, setPoints] = useState<dealerPointsApi.DealerPoint[]>([])
  const [legalEntities, setLegalEntities] = useState<dealerPointsApi.LegalEntity[]>([])
  const [warehouses, setWarehouses] = useState<dealerPointsApi.Warehouse[]>([])
  const [form, setForm] = useState<VehicleFormType>({
    vin: '',
    make: '',
    model: '',
    year: new Date().getFullYear(),
    mileage_km: 0,
    price: '',
    status: 'available',
    color: '',
    notes: '',
    brand_id: undefined,
    dealer_point_id: undefined,
    legal_entity_id: undefined,
    warehouse_id: undefined,
  })

  useEffect(() => {
    brandsApi.listBrands({ limit: 500 }).then((r) => setBrands(r.brands)).catch(() => setBrands([]))
    dealerPointsApi.listDealerPoints({ limit: 200 }).then((r) => setPoints(r.dealer_points)).catch(() => setPoints([]))
  }, [])

  useEffect(() => {
    if (!form.dealer_point_id) {
      setLegalEntities([])
      setWarehouses([])
      return
    }
    dealerPointsApi.listLegalEntitiesByDealerPoint(form.dealer_point_id).then(setLegalEntities).catch(() => setLegalEntities([]))
  }, [form.dealer_point_id])

  useEffect(() => {
    if (!form.dealer_point_id || !form.legal_entity_id) {
      setWarehouses([])
      return
    }
    dealerPointsApi
      .listWarehouses({ limit: 200, dealer_point_id: form.dealer_point_id, legal_entity_id: form.legal_entity_id, type: 'cars' })
      .then((r) => setWarehouses(r.warehouses))
      .catch(() => setWarehouses([]))
  }, [form.dealer_point_id, form.legal_entity_id])

  useEffect(() => {
    if (isNew) return
    api.getVehicle(id!)
      .then((v) => {
        setForm({
          vin: v.vin,
          make: v.make || '',
          model: v.model || '',
          year: v.year,
          mileage_km: v.mileage_km ?? 0,
          price: v.price || '',
          status: v.status || 'available',
          color: v.color || '',
          notes: v.notes || '',
          brand_id: v.brand_id || undefined,
          dealer_point_id: v.dealer_point_id || undefined,
          legal_entity_id: v.legal_entity_id || undefined,
          warehouse_id: v.warehouse_id || undefined,
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
      vin: form.vin,
      make: form.make || undefined,
      model: form.model || undefined,
      year: form.year,
      mileage_km: form.mileage_km,
      price: form.price || undefined,
      status: form.status || undefined,
      color: form.color || undefined,
      notes: form.notes || undefined,
      brand_id: form.brand_id ?? '',
      dealer_point_id: form.dealer_point_id ?? '',
      legal_entity_id: form.legal_entity_id ?? '',
      warehouse_id: form.warehouse_id ?? '',
    }
    if (isNew) {
      api.createVehicle(payload)
        .then(() => navigate('/vehicles', { replace: true }))
        .catch((err) => { setError(err.message); setSubmitting(false) })
    } else {
      api.updateVehicle(id!, payload)
        .then(() => navigate('/vehicles', { replace: true }))
        .catch((err) => { setError(err.message); setSubmitting(false) })
    }
  }

  if (loading) return <div className="main loading">Загрузка…</div>

  return (
    <div className="form-card">
      <h1 className="form-title">{isNew ? 'Новый автомобиль' : 'Редактирование автомобиля'}</h1>
      <form onSubmit={handleSubmit} className="form">
        {error && <div className="form-error">{error}</div>}
        <label className="form-label">
          VIN *
          <input
            value={form.vin}
            onChange={(e) => setForm((f) => ({ ...f, vin: e.target.value }))}
            required
            placeholder="WVWZZZ3CZWE123456"
            className="form-input"
          />
        </label>
        <label className="form-label">
          Бренд
          <select
            value={form.brand_id ?? ''}
            onChange={(e) => setForm((f) => ({ ...f, brand_id: e.target.value || undefined }))}
            className="form-input"
            style={{ maxWidth: '280px' }}
          >
            <option value="">— не выбран —</option>
            {brands.map((b) => (
              <option key={b.id} value={b.id}>{b.name}</option>
            ))}
          </select>
        </label>
        <div style={{ display: 'flex', gap: '12px', flexWrap: 'wrap' }}>
          <label className="form-label" style={{ flex: '1 1 200px' }}>
            Дилерская точка
            <select
              value={form.dealer_point_id ?? ''}
              onChange={(e) => setForm((f) => ({ ...f, dealer_point_id: e.target.value || undefined, legal_entity_id: undefined, warehouse_id: undefined }))}
              className="form-input"
            >
              <option value="">— не выбрана —</option>
              {points.map((p) => (
                <option key={p.id} value={p.id}>{p.name}</option>
              ))}
            </select>
          </label>
          <label className="form-label" style={{ flex: '1 1 200px' }}>
            Юр. лицо
            <select
              value={form.legal_entity_id ?? ''}
              onChange={(e) => setForm((f) => ({ ...f, legal_entity_id: e.target.value || undefined, warehouse_id: undefined }))}
              className="form-input"
              disabled={!form.dealer_point_id}
            >
              <option value="">— не выбрано —</option>
              {legalEntities.map((e) => (
                <option key={e.id} value={e.id}>{e.name}</option>
              ))}
            </select>
          </label>
          <label className="form-label" style={{ flex: '1 1 200px' }}>
            Склад автомобилей
            <select
              value={form.warehouse_id ?? ''}
              onChange={(e) => setForm((f) => ({ ...f, warehouse_id: e.target.value || undefined }))}
              className="form-input"
              disabled={!form.legal_entity_id}
            >
              <option value="">— не выбран —</option>
              {warehouses.map((w) => (
                <option key={w.id} value={w.id}>{w.name}</option>
              ))}
            </select>
          </label>
        </div>
        <div style={{ display: 'flex', gap: '12px', flexWrap: 'wrap' }}>
          <label className="form-label" style={{ flex: '1 1 200px' }}>
            Марка
            <input
              value={form.make}
              onChange={(e) => setForm((f) => ({ ...f, make: e.target.value }))}
              className="form-input"
            />
          </label>
          <label className="form-label" style={{ flex: '1 1 200px' }}>
            Модель
            <input
              value={form.model}
              onChange={(e) => setForm((f) => ({ ...f, model: e.target.value }))}
              className="form-input"
            />
          </label>
        </div>
        <div style={{ display: 'flex', gap: '12px', flexWrap: 'wrap' }}>
          <label className="form-label">
            Год
            <input
              type="number"
              min={1900}
              max={2100}
              value={form.year ?? ''}
              onChange={(e) => setForm((f) => ({ ...f, year: e.target.value ? parseInt(e.target.value, 10) : undefined }))}
              className="form-input"
              style={{ width: '100px' }}
            />
          </label>
          <label className="form-label">
            Пробег (км)
            <input
              type="number"
              min={0}
              value={form.mileage_km ?? ''}
              onChange={(e) => setForm((f) => ({ ...f, mileage_km: e.target.value ? parseInt(e.target.value, 10) : 0 }))}
              className="form-input"
              style={{ width: '120px' }}
            />
          </label>
          <label className="form-label">
            Цена
            <input
              value={form.price}
              onChange={(e) => setForm((f) => ({ ...f, price: e.target.value }))}
              className="form-input"
              style={{ width: '140px' }}
            />
          </label>
          <label className="form-label">
            Статус
            <select
              value={form.status}
              onChange={(e) => setForm((f) => ({ ...f, status: e.target.value }))}
              className="form-input"
              style={{ width: '160px' }}
            >
              <option value="available">В наличии</option>
              <option value="reserved">Зарезервирован</option>
              <option value="sold">Продан</option>
            </select>
          </label>
        </div>
        <label className="form-label">
          Цвет
          <input
            value={form.color}
            onChange={(e) => setForm((f) => ({ ...f, color: e.target.value }))}
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
          <button type="button" onClick={() => navigate('/vehicles')} className="form-cancel">
            Отмена
          </button>
        </div>
      </form>
    </div>
  )
}
