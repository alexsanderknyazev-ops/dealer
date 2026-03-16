import { useEffect, useState } from 'react'
import { useNavigate, useParams, useSearchParams } from 'react-router-dom'
import type { PartForm as PartFormType, PartStockRow } from './partsApi'
import * as api from './partsApi'
import * as brandsApi from './brandsApi'
import * as dealerPointsApi from './dealerPointsApi'
import './Form.css'

export function PartForm() {
  const { id } = useParams()
  const [searchParams] = useSearchParams()
  const navigate = useNavigate()
  const isNew = id === 'new' || !id
  const defaultFolderId = searchParams.get('folder_id') || ''
  const [loading, setLoading] = useState(!isNew)
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)
  const [folderOptions, setFolderOptions] = useState<{ id: string; name: string; level: number }[]>([])
  const [brands, setBrands] = useState<brandsApi.Brand[]>([])
  const [points, setPoints] = useState<dealerPointsApi.DealerPoint[]>([])
  const [legalEntities, setLegalEntities] = useState<dealerPointsApi.LegalEntity[]>([])
  const [warehouses, setWarehouses] = useState<dealerPointsApi.Warehouse[]>([])
  const [allPartsWarehouses, setAllPartsWarehouses] = useState<dealerPointsApi.Warehouse[]>([])
  const [form, setForm] = useState<PartFormType>({
    sku: '',
    name: '',
    category: '',
    folder_id: defaultFolderId || undefined,
    brand_id: undefined,
    dealer_point_id: undefined,
    legal_entity_id: undefined,
    warehouse_id: undefined,
    quantity: 0,
    unit: 'шт',
    price: '',
    location: '',
    notes: '',
    stock: [],
  })

  useEffect(() => {
    api.loadAllFoldersFlat().then(setFolderOptions).catch(() => setFolderOptions([]))
    brandsApi.listBrands({ limit: 500 }).then((r) => setBrands(r.brands)).catch(() => setBrands([]))
    dealerPointsApi.listDealerPoints({ limit: 200 }).then((r) => setPoints(r.dealer_points)).catch(() => setPoints([]))
    dealerPointsApi.listWarehouses({ limit: 200, type: 'parts' }).then((r) => setAllPartsWarehouses(r.warehouses)).catch(() => setAllPartsWarehouses([]))
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
      .listWarehouses({ limit: 200, dealer_point_id: form.dealer_point_id, legal_entity_id: form.legal_entity_id, type: 'parts' })
      .then((r) => setWarehouses(r.warehouses))
      .catch(() => setWarehouses([]))
  }, [form.dealer_point_id, form.legal_entity_id])

  useEffect(() => {
    if (isNew) {
      if (defaultFolderId) setForm((f) => ({ ...f, folder_id: defaultFolderId }))
      return
    }
    api.getPart(id!)
      .then((p) => {
        setForm({
          sku: p.sku,
          name: p.name || '',
          category: p.category || '',
          folder_id: p.folder_id || undefined,
          brand_id: p.brand_id || undefined,
          dealer_point_id: p.dealer_point_id || undefined,
          legal_entity_id: p.legal_entity_id || undefined,
          warehouse_id: p.warehouse_id || undefined,
          quantity: p.quantity ?? 0,
          unit: p.unit || 'шт',
          price: p.price || '',
          location: p.location || '',
          notes: p.notes || '',
          stock: p.stock && p.stock.length > 0 ? p.stock : [],
        })
      })
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false))
  }, [id, isNew, defaultFolderId])

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!form.sku?.trim()) {
      setError('Укажите артикул')
      return
    }
    setError(null)
    setSubmitting(true)
    const stockFiltered = (form.stock ?? []).filter((s) => s.warehouse_id && s.quantity >= 0)
    const payload: PartFormType & { stock?: PartStockRow[] } = {
      sku: form.sku.trim(),
      name: form.name || undefined,
      category: form.category || undefined,
      folder_id: form.folder_id || undefined,
      brand_id: form.brand_id ?? '',
      dealer_point_id: form.dealer_point_id ?? '',
      legal_entity_id: form.legal_entity_id ?? '',
      warehouse_id: form.warehouse_id ?? '',
      quantity: form.quantity,
      unit: form.unit || undefined,
      price: form.price || undefined,
      location: form.location || undefined,
      notes: form.notes || undefined,
    }
    if (stockFiltered.length > 0) {
      payload.stock = stockFiltered
    }
    const returnUrl = form.folder_id ? `/parts?folder_id=${form.folder_id}` : '/parts'
    if (isNew) {
      api.createPart(payload)
        .then(() => navigate(returnUrl, { replace: true }))
        .catch((err) => { setError(err.message); setSubmitting(false) })
    } else {
      api.updatePart(id!, payload)
        .then(() => navigate(returnUrl, { replace: true }))
        .catch((err) => { setError(err.message); setSubmitting(false) })
    }
  }

  if (loading) return <div className="main loading">Загрузка…</div>

  return (
    <div className="form-card">
      <h1 className="form-title">{isNew ? 'Новая запчасть' : 'Редактирование запчасти'}</h1>
      <form onSubmit={handleSubmit} className="form">
        {error && <div className="form-error">{error}</div>}
        <label className="form-label">
          Артикул (SKU) *
          <input
            value={form.sku}
            onChange={(e) => setForm((f) => ({ ...f, sku: e.target.value }))}
            required
            placeholder="ABC-12345"
            className="form-input"
          />
        </label>
        <label className="form-label">
          Папка
          <select
            value={form.folder_id ?? ''}
            onChange={(e) => setForm((f) => ({ ...f, folder_id: e.target.value || undefined }))}
            className="form-input"
          >
            <option value="">Без папки</option>
            {folderOptions.map((opt) => (
              <option key={opt.id} value={opt.id}>
                {'—'.repeat(opt.level)} {opt.name}
              </option>
            ))}
          </select>
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
            Склад запчастей
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
        <div className="form-label" style={{ marginTop: 16 }}>
          <strong>Остатки по складам</strong> (запчасть может быть на нескольких складах)
        </div>
        {(form.stock ?? []).map((row, idx) => (
          <div key={idx} style={{ display: 'flex', gap: 8, alignItems: 'center', flexWrap: 'wrap' }}>
            <select
              value={row.warehouse_id}
              onChange={(e) =>
                setForm((f) => ({
                  ...f,
                  stock: (f.stock ?? []).map((s, i) => (i === idx ? { ...s, warehouse_id: e.target.value } : s)),
                }))
              }
              className="form-input"
              style={{ minWidth: 220 }}
            >
              <option value="">— выберите склад —</option>
              {allPartsWarehouses.map((w) => (
                <option key={w.id} value={w.id}>{w.name}</option>
              ))}
            </select>
            <input
              type="number"
              min={0}
              value={row.quantity}
              onChange={(e) =>
                setForm((f) => ({
                  ...f,
                  stock: (f.stock ?? []).map((s, i) => (i === idx ? { ...s, quantity: parseInt(e.target.value, 10) || 0 } : s)),
                }))
              }
              className="form-input"
              style={{ width: 80 }}
            />
            <span>{form.unit || 'шт'}</span>
            <button
              type="button"
              onClick={() => setForm((f) => ({ ...f, stock: (f.stock ?? []).filter((_, i) => i !== idx) }))}
              className="form-cancel"
              style={{ padding: '4px 8px' }}
            >
              Удалить
            </button>
          </div>
        ))}
        <button
          type="button"
          onClick={() => setForm((f) => ({ ...f, stock: [...(f.stock ?? []), { warehouse_id: '', quantity: 0 }] }))}
          className="form-cancel"
          style={{ marginTop: 4 }}
        >
          + Добавить склад
        </button>
        <label className="form-label">
          Название
          <input
            value={form.name}
            onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))}
            className="form-input"
            placeholder="Масляный фильтр"
          />
        </label>
        <div style={{ display: 'flex', gap: '12px', flexWrap: 'wrap' }}>
          <label className="form-label" style={{ flex: '1 1 200px' }}>
            Категория
            <input
              value={form.category}
              onChange={(e) => setForm((f) => ({ ...f, category: e.target.value }))}
              className="form-input"
              placeholder="Фильтры, Тормоза..."
            />
          </label>
          <label className="form-label">
            Количество
            <input
              type="number"
              min={0}
              value={form.quantity ?? ''}
              onChange={(e) => setForm((f) => ({ ...f, quantity: e.target.value ? parseInt(e.target.value, 10) : 0 }))}
              className="form-input"
              style={{ width: '100px' }}
            />
          </label>
          <label className="form-label">
            Ед. изм.
            <select
              value={form.unit}
              onChange={(e) => setForm((f) => ({ ...f, unit: e.target.value }))}
              className="form-input"
              style={{ width: '120px' }}
            >
              <option value="шт">шт</option>
              <option value="комплект">комплект</option>
              <option value="л">л</option>
              <option value="кг">кг</option>
            </select>
          </label>
          <label className="form-label">
            Цена
            <input
              value={form.price}
              onChange={(e) => setForm((f) => ({ ...f, price: e.target.value }))}
              className="form-input"
              style={{ width: '120px' }}
            />
          </label>
        </div>
        <label className="form-label">
          Расположение (склад/полка)
          <input
            value={form.location}
            onChange={(e) => setForm((f) => ({ ...f, location: e.target.value }))}
            className="form-input"
            placeholder="Склад А, полка 12"
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
          <button type="button" onClick={() => navigate(form.folder_id ? `/parts?folder_id=${form.folder_id}` : '/parts')} className="form-cancel">
            Отмена
          </button>
        </div>
      </form>
    </div>
  )
}
