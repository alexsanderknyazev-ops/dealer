import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import type { Vehicle } from './vehiclesApi'
import * as api from './vehiclesApi'
import './Vehicles.css'

const STATUS_LABEL: Record<string, string> = {
  available: 'В наличии',
  sold: 'Продан',
  reserved: 'Зарезервирован',
}

export function Vehicles() {
  const [list, setList] = useState<Vehicle[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [search, setSearch] = useState('')
  const [statusFilter, setStatusFilter] = useState('')
  const [page, setPage] = useState(0)
  const limit = 20

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    setError(null)
    api.listVehicles({ limit, offset: page * limit, search: search || undefined, status: statusFilter || undefined })
      .then((r) => {
        if (!cancelled) {
          setList(r.vehicles)
          setTotal(r.total)
        }
      })
      .catch((err) => {
        if (!cancelled) {
          setList([])
          setError(err instanceof Error ? err.message : 'Ошибка загрузки')
        }
      })
      .finally(() => { if (!cancelled) setLoading(false) })
    return () => { cancelled = true }
  }, [page, search, statusFilter])

  return (
    <div className="vehicles">
      <div className="vehicles-header">
        <h1 className="vehicles-title">Автомобили</h1>
        <Link to="/vehicles/new" className="vehicles-add">+ Добавить</Link>
      </div>
      <div className="vehicles-toolbar">
        <input
          type="search"
          placeholder="Поиск по VIN, марке, модели..."
          value={search}
          onChange={(e) => { setSearch(e.target.value); setPage(0) }}
          className="vehicles-search"
        />
        <select
          value={statusFilter}
          onChange={(e) => { setStatusFilter(e.target.value); setPage(0) }}
          className="vehicles-status-filter"
        >
          <option value="">Все статусы</option>
          <option value="available">В наличии</option>
          <option value="reserved">Зарезервирован</option>
          <option value="sold">Продан</option>
        </select>
      </div>
      {error && (
        <div className="vehicles-error">
          {error}. Запустите vehicles-service: <code>make run-vehicles</code>
        </div>
      )}
      {loading ? (
        <p className="vehicles-loading">Загрузка…</p>
      ) : list.length === 0 && !error ? (
        <p className="vehicles-empty">Нет автомобилей</p>
      ) : (
        <>
          <table className="vehicles-table">
            <thead>
              <tr>
                <th>VIN</th>
                <th>Марка / Модель</th>
                <th>Год</th>
                <th>Пробег</th>
                <th>Цена</th>
                <th>Статус</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              {list.map((v) => (
                <tr key={v.id}>
                  <td className="vehicles-vin">{v.vin}</td>
                  <td>{v.make} {v.model}</td>
                  <td>{v.year}</td>
                  <td>{v.mileage_km.toLocaleString('ru')} км</td>
                  <td>{v.price ? Number(v.price).toLocaleString('ru') : '—'}</td>
                  <td>{STATUS_LABEL[v.status] || v.status}</td>
                  <td>
                    <Link to={`/vehicles/${v.id}`} className="vehicles-link">Открыть</Link>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
          {total > limit && (
            <div className="vehicles-pagination">
              <button type="button" disabled={page === 0} onClick={() => setPage((p) => p - 1)}>Назад</button>
              <span>Стр. {page + 1} из {Math.ceil(total / limit) || 1}</span>
              <button type="button" disabled={(page + 1) * limit >= total} onClick={() => setPage((p) => p + 1)}>Вперёд</button>
            </div>
          )}
        </>
      )}
    </div>
  )
}
