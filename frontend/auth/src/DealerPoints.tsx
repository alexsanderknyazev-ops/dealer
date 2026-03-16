import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import type { DealerPoint } from './dealerPointsApi'
import * as api from './dealerPointsApi'
import './Customers.css'

export function DealerPoints() {
  const [list, setList] = useState<DealerPoint[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [search, setSearch] = useState('')
  const [page, setPage] = useState(0)
  const limit = 20

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    setError(null)
    api.listDealerPoints({ limit, offset: page * limit, search: search || undefined })
      .then((r) => {
        if (!cancelled) {
          setList(r.dealer_points)
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
  }, [page, search])

  return (
    <div className="customers">
      <div className="customers-header">
        <h1 className="customers-title">Дилерские точки</h1>
        <Link to="/dealer-points/new" className="customers-add">+ Добавить</Link>
      </div>
      <div className="customers-toolbar">
        <input
          type="search"
          placeholder="Поиск по названию, адресу..."
          value={search}
          onChange={(e) => { setSearch(e.target.value); setPage(0) }}
          className="customers-search"
        />
      </div>
      {error && (
        <div className="customers-error">
          <p style={{ margin: '0 0 8px 0' }}>{error}</p>
        </div>
      )}
      {loading ? (
        <p className="customers-loading">Загрузка…</p>
      ) : list.length === 0 && !error ? (
        <p className="customers-empty">Нет дилерских точек. Нажмите «+ Добавить».</p>
      ) : (
        <>
          <table className="customers-table">
            <thead>
              <tr>
                <th>Название</th>
                <th>Адрес</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              {list.map((d) => (
                <tr key={d.id}>
                  <td>{d.name}</td>
                  <td>{d.address || '—'}</td>
                  <td style={{ whiteSpace: 'nowrap' }}>
                    <Link to={`/dealer-points/${d.id}/edit`} className="customers-link">Изменить</Link>
                    {' · '}
                    <Link to={`/legal-entities?dealer_point_id=${d.id}`} className="customers-link">Юр. лица</Link>
                    {' · '}
                    <Link to={`/warehouses?dealer_point_id=${d.id}`} className="customers-link">Склады</Link>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
          {total > limit && (
            <div className="customers-pagination">
              <span>Стр. {page + 1} из {Math.ceil(total / limit) || 1}</span>
              <button type="button" disabled={page === 0} onClick={() => setPage((p) => p - 1)}>Назад</button>
              <button type="button" disabled={(page + 1) * limit >= total} onClick={() => setPage((p) => p + 1)}>Вперёд</button>
            </div>
          )}
        </>
      )}
    </div>
  )
}
