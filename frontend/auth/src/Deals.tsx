import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import type { Deal } from './dealsApi'
import * as api from './dealsApi'
import './Deals.css'

const STAGE_LABEL: Record<string, string> = {
  draft: 'Черновик',
  in_progress: 'В работе',
  paid: 'Оплачено',
  completed: 'Завершено',
  cancelled: 'Отменено',
}

export function Deals() {
  const [list, setList] = useState<Deal[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [stageFilter, setStageFilter] = useState('')
  const [page, setPage] = useState(0)
  const [retry, setRetry] = useState(0)
  const limit = 20

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    setError(null)
    api.listDeals({ limit, offset: page * limit, stage: stageFilter || undefined })
      .then((r) => {
        if (!cancelled) {
          setList(r.deals)
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
  }, [page, stageFilter, retry])

  return (
    <div className="deals">
      <div className="deals-header">
        <h1 className="deals-title">Сделки</h1>
        <Link to="/deals/new" className="deals-add">+ Новая сделка</Link>
      </div>
      <div className="deals-toolbar">
        <select
          value={stageFilter}
          onChange={(e) => { setStageFilter(e.target.value); setPage(0) }}
          className="deals-stage-filter"
        >
          <option value="">Все этапы</option>
          <option value="draft">Черновик</option>
          <option value="in_progress">В работе</option>
          <option value="paid">Оплачено</option>
          <option value="completed">Завершено</option>
          <option value="cancelled">Отменено</option>
        </select>
      </div>
      {error && (
        <div className="deals-error">
          <p style={{ margin: '0 0 8px 0' }}>{error}</p>
          <button type="button" onClick={() => setRetry((r) => r + 1)} className="deals-retry">Повторить</button>
        </div>
      )}
      {loading ? (
        <p className="deals-loading">Загрузка…</p>
      ) : list.length === 0 && !error ? (
        <p className="deals-empty">
          Нет сделок. Нажмите «+ Новая сделка» для создания.
        </p>
      ) : (
        <>
          <table className="deals-table">
            <thead>
              <tr>
                <th>Клиент ID</th>
                <th>Авто ID</th>
                <th>Сумма</th>
                <th>Этап</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              {list.map((d) => (
                <tr key={d.id}>
                  <td className="deals-id">{d.customer_id.slice(0, 8)}…</td>
                  <td className="deals-id">{d.vehicle_id.slice(0, 8)}…</td>
                  <td>{d.amount ? Number(d.amount).toLocaleString('ru') : '—'}</td>
                  <td>{STAGE_LABEL[d.stage] || d.stage}</td>
                  <td>
                    <Link to={`/deals/${d.id}`} className="deals-link">Открыть</Link>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
          {total > limit && (
            <div className="deals-pagination">
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
