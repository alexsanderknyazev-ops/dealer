import { useEffect, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import type { Part } from './partsApi'
import * as api from './partsApi'
import * as dealerPointsApi from './dealerPointsApi'
import './CustomerView.css'

export function PartView() {
  const { id } = useParams()
  const navigate = useNavigate()
  const [part, setPart] = useState<Part | null>(null)
  const [folderName, setFolderName] = useState<string | null>(null)
  const [pointName, setPointName] = useState<string | null>(null)
  const [legalEntityName, setLegalEntityName] = useState<string | null>(null)
  const [warehouseName, setWarehouseName] = useState<string | null>(null)
  const [warehouseNames, setWarehouseNames] = useState<Record<string, string>>({})
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!id) return
    api.getPart(id)
      .then((p) => {
        setPart(p)
        if (p.folder_id) {
          api.getFolder(p.folder_id).then((f) => setFolderName(f.name)).catch(() => {})
        }
        if (p.dealer_point_id) {
          dealerPointsApi.getDealerPoint(p.dealer_point_id).then((d) => setPointName(d.name)).catch(() => {})
        }
        if (p.legal_entity_id) {
          dealerPointsApi.getLegalEntity(p.legal_entity_id).then((e) => setLegalEntityName(e.name)).catch(() => {})
        }
        if (p.warehouse_id) {
          dealerPointsApi.getWarehouse(p.warehouse_id).then((w) => setWarehouseName(w.name)).catch(() => {})
        }
        if (p.stock && p.stock.length > 0) {
          const ids = [...new Set(p.stock.map((s) => s.warehouse_id))]
          Promise.all(ids.map((wid) => dealerPointsApi.getWarehouse(wid).then((w) => ({ id: wid, name: w.name }))))
            .then((pairs) => setWarehouseNames(Object.fromEntries(pairs.map(({ id: i, name }) => [i, name]))))
            .catch(() => {})
        }
      })
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false))
  }, [id])

  const hasLocation = part?.dealer_point_id || part?.legal_entity_id || part?.warehouse_id
  const hasStock = part?.stock && part.stock.length > 0

  async function handleDelete() {
    if (!id || !part || !confirm(`Удалить запчасть ${part.sku} — ${part.name || 'без названия'}?`)) return
    try {
      await api.deletePart(id)
      navigate('/parts', { replace: true })
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Ошибка удаления')
    }
  }

  if (loading) return <div className="main loading">Загрузка…</div>
  if (error || !part) return <div className="form-error">{error || 'Не найдено'}</div>

  return (
    <div className="customer-view">
      <div className="customer-view-header">
        <h1 className="customer-view-name">{part.name || part.sku}</h1>
        <div className="customer-view-actions">
          <Link to={`/parts/${id}/edit`} className="customer-view-btn customer-view-edit">Редактировать</Link>
          <button type="button" onClick={handleDelete} className="customer-view-btn customer-view-delete">Удалить</button>
        </div>
      </div>
      <dl className="customer-view-dl">
        <dt>Артикул (SKU)</dt>
        <dd style={{ fontFamily: 'monospace' }}>{part.sku}</dd>
        <dt>Название</dt>
        <dd>{part.name || '—'}</dd>
        <dt>Категория</dt>
        <dd>{part.category || '—'}</dd>
        {folderName != null && (
          <>
            <dt>Папка</dt>
            <dd>{folderName}</dd>
          </>
        )}
        <dt>Количество</dt>
        <dd>{part.quantity} {part.unit || 'шт'}{hasStock && ` (всего по складам)`}</dd>
        <dt>Цена</dt>
        <dd>{part.price ? Number(part.price).toLocaleString('ru') : '—'}</dd>
        {part.location && (
          <>
            <dt>Расположение</dt>
            <dd>{part.location}</dd>
          </>
        )}
        {part.notes && (
          <>
            <dt>Заметки</dt>
            <dd>{part.notes}</dd>
          </>
        )}
        {hasStock && (
          <>
            <dt>Наличие по складам</dt>
            <dd className="customer-view-location">
              <ul style={{ margin: 0, paddingLeft: 20 }}>
                {part.stock!.map((s) => (
                  <li key={s.warehouse_id}>
                    <strong>{warehouseNames[s.warehouse_id] ?? s.warehouse_id}</strong>: {s.quantity} {part.unit || 'шт'}
                  </li>
                ))}
              </ul>
              <p style={{ marginTop: 8, marginBottom: 0, fontSize: '0.95em', color: 'var(--color-muted, #666)' }}>
                Запчасть может быть на нескольких складах. Изменить остатки — <Link to={`/parts/${id}/edit`}>Редактировать</Link>.
              </p>
            </dd>
          </>
        )}
        {hasLocation && !hasStock && (
          <>
            <dt>Расположение (склад)</dt>
            <dd className="customer-view-location">
              <div><strong>Дилерская точка:</strong> {pointName ?? part.dealer_point_id ?? '—'}</div>
              <div><strong>Юр. лицо:</strong> {legalEntityName ?? part.legal_entity_id ?? '—'}</div>
              <div><strong>Склад запчастей:</strong> {warehouseName ?? (part.warehouse_id ? `склад ${part.warehouse_id.slice(0, 8)}…` : '—')}</div>
              <p style={{ marginTop: 8, marginBottom: 0, fontSize: '0.95em', color: 'var(--color-muted, #666)' }}>
                Чтобы указать несколько складов или изменить остатки — <Link to={`/parts/${id}/edit`}>Редактировать</Link>.
              </p>
            </dd>
          </>
        )}
      </dl>
      <p className="customer-view-back">
        <Link to="/parts">← К списку запчастей</Link>
      </p>
    </div>
  )
}
