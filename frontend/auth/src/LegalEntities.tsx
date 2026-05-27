import { useEffect, useState } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import type { LegalEntity } from './dealerPointsApi'
import * as api from './dealerPointsApi'
import './Customers.css'

export function LegalEntities() {
  const [searchParams, setSearchParams] = useSearchParams()
  const dealerPointId = searchParams.get('dealer_point_id') || ''
  const [list, setList] = useState<LegalEntity[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [search, setSearch] = useState('')
  const [page, setPage] = useState(0)
  const [points, setPoints] = useState<{ id: string; name: string }[]>([])
  const limit = 20

  useEffect(() => {
    api.listDealerPoints({ limit: 200 }).then((r) => setPoints(r.dealer_points)).catch(() => setPoints([]))
  }, [])

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    setError(null)
    if (dealerPointId) {
      api.listLegalEntitiesByDealerPoint(dealerPointId)
        .then((linked) => {
          if (!cancelled) {
            const filtered = search
              ? linked.filter((e) =>
                  e.name.toLowerCase().includes(search.toLowerCase()) ||
                  (e.inn && e.inn.includes(search)))
              : linked
            setList(filtered.slice(page * limit, (page + 1) * limit))
            setTotal(filtered.length)
          }
        })
        .catch((err) => {
          if (!cancelled) {
            setList([])
            setError(err instanceof Error ? err.message : 'Ошибка загрузки')
          }
        })
        .finally(() => { if (!cancelled) setLoading(false) })
    } else {
      api.listLegalEntities({ limit, offset: page * limit, search: search || undefined })
        .then((r) => {
          if (!cancelled) {
            setList(r.legal_entities)
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
    }
    return () => { cancelled = true }
  }, [page, search, dealerPointId])

  const pointName = points.find((p) => p.id === dealerPointId)?.name

  return (
    <div className="customers">
      <div className="customers-header">
        <h1 className="customers-title">
          Юридические лица
          {pointName && <span style={{ fontWeight: 'normal', fontSize: '0.9em' }}> — {pointName}</span>}
        </h1>
        <Link to="/legal-entities/new" className="customers-add">+ Добавить</Link>
      </div>
      <div className="customers-toolbar">
        {dealerPointId && (
          <Link to="/legal-entities" className="customers-link" style={{ marginRight: 12 }}>Все юр. лица</Link>
        )}
        {points.length > 0 && (
          <select
            value={dealerPointId}
            onChange={(e) => {
              const v = e.target.value
              setPage(0)
              const next = new URLSearchParams(searchParams)
              if (v) next.set('dealer_point_id', v)
              else next.delete('dealer_point_id')
              setSearchParams(next)
            }}
            className="customers-search"
            style={{ maxWidth: 280 }}
          >
            <option value="">Все дилерские точки</option>
            {points.map((p) => (
              <option key={p.id} value={p.id}>{p.name}</option>
            ))}
          </select>
        )}
        <input
          type="search"
          placeholder="Поиск по названию, ИНН..."
          value={search}
          onChange={(e) => { setSearch(e.target.value); setPage(0) }}
          className="customers-search"
          style={{ marginLeft: 8 }}
        />
      </div>
      {error && <div className="customers-error">{error}</div>}
      {loading ? (
        <p className="customers-loading">Загрузка…</p>
      ) : list.length === 0 && !error ? (
        <p className="customers-empty">Нет юридических лиц.</p>
      ) : (
        <>
          <table className="customers-table">
            <thead>
              <tr>
                <th>Название</th>
                <th>ИНН</th>
                <th>Адрес</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              {list.map((e) => (
                <tr key={e.id}>
                  <td>{e.name}</td>
                  <td>{e.inn || '—'}</td>
                  <td>{e.address || '—'}</td>
                  <td style={{ whiteSpace: 'nowrap' }}>
                    <Link to={`/legal-entities/${e.id}/edit`} className="customers-link">Изменить</Link>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
          {total > limit && (
            <div className="customers-pagination">
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
