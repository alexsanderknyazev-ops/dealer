import { useEffect, useState } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import type { Warehouse } from './dealerPointsApi'
import * as api from './dealerPointsApi'
import './Customers.css'

const TYPE_LABEL: Record<string, string> = {
  cars: 'Автомобили',
  parts: 'Запчасти',
}

export function Warehouses() {
  const [searchParams, setSearchParams] = useSearchParams()
  const typeFilter = (searchParams.get('type') as 'cars' | 'parts' | '') || ''
  const dealerPointId = searchParams.get('dealer_point_id') || ''
  const [list, setList] = useState<Warehouse[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
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
    api
      .listWarehouses({
        limit,
        offset: page * limit,
        dealer_point_id: dealerPointId || undefined,
        type: typeFilter === 'cars' || typeFilter === 'parts' ? typeFilter : undefined,
      })
      .then((r) => {
        if (!cancelled) {
          setList(r.warehouses)
          setTotal(r.total)
        }
      })
      .catch((err) => {
        if (!cancelled) {
          setList([])
          setError(err instanceof Error ? err.message : 'Ошибка загрузки')
        }
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [page, dealerPointId, typeFilter])

  return (
    <div className="customers">
      <div className="customers-header">
        <h1 className="customers-title">Склады</h1>
        <Link to="/warehouses/new" className="customers-add">
          + Добавить
        </Link>
      </div>
      <div className="customers-toolbar" style={{ flexWrap: 'wrap', gap: 8 }}>
        <span style={{ marginRight: 4 }}>Тип:</span>
        <Link
          to={dealerPointId ? `/warehouses?dealer_point_id=${dealerPointId}` : '/warehouses'}
          style={{
            padding: '4px 8px',
            borderRadius: 4,
            background: !typeFilter ? 'var(--color-bg, #eee)' : undefined,
            textDecoration: 'none',
            color: 'inherit',
          }}
        >
          Все
        </Link>
        <Link
          to={
            dealerPointId
              ? `/warehouses?type=cars&dealer_point_id=${dealerPointId}`
              : '/warehouses?type=cars'
          }
          style={{
            padding: '4px 8px',
            borderRadius: 4,
            background: typeFilter === 'cars' ? 'var(--color-bg, #eee)' : undefined,
            textDecoration: 'none',
            color: 'inherit',
          }}
        >
          Склады автомобилей
        </Link>
        <Link
          to={
            dealerPointId
              ? `/warehouses?type=parts&dealer_point_id=${dealerPointId}`
              : '/warehouses?type=parts'
          }
          style={{
            padding: '4px 8px',
            borderRadius: 4,
            background: typeFilter === 'parts' ? 'var(--color-bg, #eee)' : undefined,
            textDecoration: 'none',
            color: 'inherit',
          }}
        >
          Склады запчастей
        </Link>
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
            style={{ maxWidth: 280, marginLeft: 16 }}
          >
            <option value="">Все дилерские точки</option>
            {points.map((p) => (
              <option key={p.id} value={p.id}>
                {p.name}
              </option>
            ))}
          </select>
        )}
      </div>
      {error && (
        <div className="customers-error">
          <p style={{ margin: '0 0 8px 0' }}>{error}</p>
        </div>
      )}
      {loading ? (
        <p className="customers-loading">Загрузка…</p>
      ) : list.length === 0 && !error ? (
        <p className="customers-empty">Нет складов. Нажмите «+ Добавить» или смените фильтры.</p>
      ) : (
        <>
          <table className="customers-table">
            <thead>
              <tr>
                <th>Название</th>
                <th>Тип</th>
                <th>Дилерская точка</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              {list.map((w) => (
                <tr key={w.id}>
                  <td>{w.name}</td>
                  <td>{TYPE_LABEL[w.type] || w.type}</td>
                  <td>
                    {points.find((p) => p.id === w.dealer_point_id)?.name ?? w.dealer_point_id}
                  </td>
                  <td style={{ whiteSpace: 'nowrap' }}>
                    <Link to={`/warehouses/${w.id}/edit`} className="customers-link">
                      Изменить
                    </Link>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
          {total > limit && (
            <div className="customers-pagination">
              <button type="button" disabled={page === 0} onClick={() => setPage((p) => p - 1)}>
                Назад
              </button>
              <span>
                Стр. {page + 1} из {Math.ceil(total / limit) || 1}
              </span>
              <button
                type="button"
                disabled={(page + 1) * limit >= total}
                onClick={() => setPage((p) => p + 1)}
              >
                Вперёд
              </button>
            </div>
          )}
        </>
      )}
    </div>
  )
}
