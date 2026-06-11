import { useEffect, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import type { Vehicle } from './vehiclesApi'
import * as api from './vehiclesApi'
import * as dealerPointsApi from './dealerPointsApi'
import { EntityViewPage } from '@/components/common/EntityViewPage'

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
    api
      .getVehicle(id)
      .then((v) => {
        setVehicle(v)
        if (v.dealer_point_id) {
          dealerPointsApi.getDealerPoint(v.dealer_point_id).then((d) => setPointName(d.name)).catch(() => {})
        }
        if (v.warehouse_id) {
          dealerPointsApi.getWarehouse(v.warehouse_id).then((w) => setWarehouseName(w.name)).catch(() => {})
        }
      })
      .catch((e) => setError(e instanceof Error ? e.message : 'Ошибка загрузки'))
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

  const locationValue =
    vehicle && (pointName || vehicle.dealer_point_id) ? (
      <span>
        {pointName ?? vehicle.dealer_point_id}
        {warehouseName && ` — ${warehouseName}`}
        {vehicle.warehouse_id && !warehouseName && ` — склад ${vehicle.warehouse_id.slice(0, 8)}…`}
        . Чтобы переместить на другой склад —{' '}
        <Link to={`/vehicles/${id}/edit`} className="text-primary hover:underline">
          Редактировать
        </Link>
        .
      </span>
    ) : null

  return (
    <EntityViewPage
      title={vehicle ? `${vehicle.make} ${vehicle.model} (${vehicle.year})` : 'Автомобиль'}
      backTo="/vehicles"
      backLabel="К списку автомобилей"
      editTo={id ? `/vehicles/${id}/edit` : undefined}
      onDelete={vehicle ? handleDelete : undefined}
      loading={loading}
      error={error || (!vehicle && !loading ? 'Не найден' : null)}
      fields={
        vehicle
          ? [
              { label: 'VIN', value: <span className="font-mono">{vehicle.vin}</span> },
              { label: 'Марка / Модель', value: `${vehicle.make} ${vehicle.model}` },
              { label: 'Год', value: vehicle.year },
              { label: 'Пробег', value: `${vehicle.mileage_km.toLocaleString('ru')} км` },
              { label: 'Цена', value: vehicle.price ? Number(vehicle.price).toLocaleString('ru') : '—' },
              { label: 'Статус', value: STATUS_LABEL[vehicle.status] || vehicle.status },
              ...(vehicle.color ? [{ label: 'Цвет', value: vehicle.color }] : []),
              ...(vehicle.notes ? [{ label: 'Заметки', value: vehicle.notes }] : []),
              ...(locationValue ? [{ label: 'Дилерская точка / Склад', value: locationValue }] : []),
            ]
          : []
      }
    />
  )
}
