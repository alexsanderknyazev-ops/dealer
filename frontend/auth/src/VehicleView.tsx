import { useEffect, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import type { Vehicle } from './vehiclesApi'
import * as api from './vehiclesApi'
import * as dealerPointsApi from './dealerPointsApi'
import './CustomerView.css'

const STATUS_LABEL: Record<string, string> = {
  available: 'В наличии',
  sold: 'Продан',
  reserved: 'Зарезервирован',
}

export function VehicleView() {
  const { id } = useParams()
  const navigate = useNavigate()
  const [vehicle, setVehicle] = useState<Vehicle | null>(null)
  const [pointName, setPointName] = useState<string | null>(null)
  const [warehouseName, setWarehouseName] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!id) return
    api.getVehicle(id)
      .then((v) => {
        setVehicle(v)
        if (v.dealer_point_id) {
          dealerPointsApi.getDealerPoint(v.dealer_point_id).then((d) => setPointName(d.name)).catch(() => {})
        }
        if (v.warehouse_id) {
          dealerPointsApi.getWarehouse(v.warehouse_id).then((w) => setWarehouseName(w.name)).catch(() => {})
        }
      })
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false))
  }, [id])

  async function handleDelete() {
    if (!id || !vehicle || !confirm(`Удалить автомобиль ${vehicle.make} ${vehicle.model} (${vehicle.vin})?`)) return
    try {
      await api.deleteVehicle(id)
      navigate('/vehicles', { replace: true })
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Ошибка удаления')
    }
  }

  if (loading) return <div className="main loading">Загрузка…</div>
  if (error || !vehicle) return <div className="form-error">{error || 'Не найден'}</div>

  return (
    <div className="customer-view">
      <div className="customer-view-header">
        <h1 className="customer-view-name">{vehicle.make} {vehicle.model} ({vehicle.year})</h1>
        <div className="customer-view-actions">
          <Link to={`/vehicles/${id}/edit`} className="customer-view-btn customer-view-edit">Редактировать</Link>
          <button type="button" onClick={handleDelete} className="customer-view-btn customer-view-delete">Удалить</button>
        </div>
      </div>
      <dl className="customer-view-dl">
        <dt>VIN</dt>
        <dd style={{ fontFamily: 'monospace' }}>{vehicle.vin}</dd>
        <dt>Марка / Модель</dt>
        <dd>{vehicle.make} {vehicle.model}</dd>
        <dt>Год</dt>
        <dd>{vehicle.year}</dd>
        <dt>Пробег</dt>
        <dd>{vehicle.mileage_km.toLocaleString('ru')} км</dd>
        <dt>Цена</dt>
        <dd>{vehicle.price ? Number(vehicle.price).toLocaleString('ru') : '—'}</dd>
        <dt>Статус</dt>
        <dd>{STATUS_LABEL[vehicle.status] || vehicle.status}</dd>
        {vehicle.color && (
          <>
            <dt>Цвет</dt>
            <dd>{vehicle.color}</dd>
          </>
        )}
        {vehicle.notes && (
          <>
            <dt>Заметки</dt>
            <dd>{vehicle.notes}</dd>
          </>
        )}
        {(pointName || vehicle.dealer_point_id) && (
          <>
            <dt>Дилерская точка / Склад</dt>
            <dd>
              {pointName ?? vehicle.dealer_point_id}
              {warehouseName && ` — ${warehouseName}`}
              {vehicle.warehouse_id && !warehouseName && ` — склад ${vehicle.warehouse_id.slice(0, 8)}…`}
              . Чтобы переместить на другой склад — <Link to={`/vehicles/${id}/edit`}>Редактировать</Link>.
            </dd>
          </>
        )}
      </dl>
      <p className="customer-view-back">
        <Link to="/vehicles">← К списку автомобилей</Link>
      </p>
    </div>
  )
}
