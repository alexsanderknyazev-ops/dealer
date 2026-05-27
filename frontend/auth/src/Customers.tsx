import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import type { Customer } from './customersApi'
import * as api from './customersApi'
import './Customers.css'

export function Customers() {
  const [list, setList] = useState<Customer[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [search, setSearch] = useState('')
  const [page, setPage] = useState(0)
  const [retry, setRetry] = useState(0)
  const limit = 20

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    setError(null)
    api.listCustomers({ limit, offset: page * limit, search: search || undefined })
      .then((r) => {
        if (!cancelled) {
          setList(r.customers)
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

  return (
    <div className="customers">
      <div className="customers-header">
        <h1 className="customers-title">Клиенты</h1>
        <Link to="/customers/new" className="customers-add">+ Добавить</Link>
      </div>
      <div className="customers-toolbar">
        <input
          type="search"
          placeholder="Поиск по имени, email, телефону..."
          value={search}
          onChange={(e) => { setSearch(e.target.value); setPage(0) }}
          className="customers-search"
        />
      </div>
      {error && (
        <div className="customers-error">
          <p style={{ margin: '0 0 8px 0' }}>{error === 'Failed to fetch'
            ? 'Нет связи с сервером. Вы сейчас на: ' + window.location.origin + ' — запросы идут на ' + window.location.origin + '/api/customers. Попробуйте открыть http://127.0.0.1:8080/customers или другой браузер.'
            : error === 'unauthorized'
              ? 'Сессия истекла. Выйдите и войдите снова.'
              : error}
          </p>
          <button type="button" onClick={() => setRetry((r) => r + 1)} className="customers-retry">Повторить</button>
        </div>
      )}
      {loading ? (
        <p className="customers-loading">Загрузка…</p>
      ) : list.length === 0 && !error ? (
        <p className="customers-empty">
          Нет клиентов. Нажмите «+ Добавить» или выполните в терминале <code>make seed-data</code> для загрузки тестовых данных.
        </p>
      ) : (
        <>
          <table className="customers-table">
            <thead>
              <tr>
                <th>Имя</th>
                <th>Email</th>
                <th>Телефон</th>
                <th>Тип</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              {list.map((c) => (
                <tr key={c.id}>
                  <td>{c.name}</td>
                  <td>{c.email || '—'}</td>
                  <td>{c.phone || '—'}</td>
                  <td>{c.customer_type === 'legal' ? 'Юр. лицо' : 'Физ. лицо'}</td>
                  <td>
                    <Link to={`/customers/${c.id}`} className="customers-link">Открыть</Link>
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
