import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import type { Brand } from './brandsApi'
import * as api from './brandsApi'
import './Customers.css'

export function Brands() {
  const [list, setList] = useState<Brand[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [search, setSearch] = useState('')
  const [page, setPage] = useState(0)
  const [retry, setRetry] = useState(0)
  const limit = 50

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    setError(null)
    api.listBrands({ limit, offset: page * limit, search: search || undefined })
      .then((r) => {
        if (!cancelled) {
          setList(r.brands)
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
  }, [page, search, retry])

  function handleDelete(id: string, name: string) {
    if (!window.confirm(`Удалить бренд «${name}»?`)) return
    api.deleteBrand(id)
      .then(() => setRetry((r) => r + 1))
      .catch((e) => setError(e.message))
  }

  return (
    <div className="customers">
      <div className="customers-header">
        <h1 className="customers-title">Бренды</h1>
        <Link to="/brands/new" className="customers-add">+ Добавить</Link>
      </div>
      <div className="customers-toolbar">
        <input
          type="search"
          placeholder="Поиск по названию..."
          value={search}
          onChange={(e) => { setSearch(e.target.value); setPage(0) }}
          className="customers-search"
        />
      </div>
      {error && (
        <div className="customers-error">
          <p style={{ margin: '0 0 8px 0' }}>{error}</p>
          <button type="button" onClick={() => setRetry((r) => r + 1)} className="customers-retry">Повторить</button>
        </div>
      )}
      {loading ? (
        <p className="customers-loading">Загрузка…</p>
      ) : list.length === 0 && !error ? (
        <p className="customers-empty">Нет брендов. Нажмите «+ Добавить».</p>
      ) : (
        <>
          <table className="customers-table">
            <thead>
              <tr>
                <th>Название</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              {list.map((b) => (
                <tr key={b.id}>
                  <td>{b.name}</td>
                  <td style={{ whiteSpace: 'nowrap' }}>
                    <Link to={`/brands/${b.id}/edit`} className="customers-link">Изменить</Link>
                    {' · '}
                    <button type="button" onClick={() => handleDelete(b.id, b.name)} className="customers-link" style={{ background: 'none', border: 'none', cursor: 'pointer', padding: 0, textDecoration: 'underline' }}>Удалить</button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
          {total > limit && (
            <div className="customers-pagination">
              <span>Всего: {total}</span>
              <button type="button" disabled={page === 0} onClick={() => setPage((p) => p - 1)}>Назад</button>
              <button type="button" disabled={(page + 1) * limit >= total} onClick={() => setPage((p) => p + 1)}>Вперёд</button>
            </div>
          )}
        </>
      )}
    </div>
  )
}
