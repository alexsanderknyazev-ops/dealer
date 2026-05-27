import { useEffect, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import type { Deal } from './dealsApi'
import * as api from './dealsApi'
import './CustomerView.css'

const STAGE_LABEL: Record<string, string> = {
  draft: 'Черновик',
  in_progress: 'В работе',
  paid: 'Оплачено',
  completed: 'Завершено',
  cancelled: 'Отменено',
}

export function DealView() {
  const { id } = useParams()
  const navigate = useNavigate()
  const [deal, setDeal] = useState<Deal | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!id) return
    api.getDeal(id)
      .then(setDeal)
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false))
  }, [id])

  async function handleDelete() {
    if (!id || !deal || !confirm('Удалить сделку?')) return
    try {
      await api.deleteDeal(id)
      navigate('/deals', { replace: true })
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Ошибка удаления')
    }
  }

  if (loading) return <div className="main loading">Загрузка…</div>
  if (error || !deal) return <div className="form-error">{error || 'Не найдено'}</div>

  return (
    <div className="customer-view">
      <div className="customer-view-header">
        <h1 className="customer-view-name">Сделка {deal.id.slice(0, 8)}…</h1>
        <div className="customer-view-actions">
          <Link to={`/deals/${id}/edit`} className="customer-view-btn customer-view-edit">Редактировать</Link>
          <button type="button" onClick={handleDelete} className="customer-view-btn customer-view-delete">Удалить</button>
        </div>
      </div>
      <dl className="customer-view-dl">
        <dt>Клиент ID</dt>
        <dd><Link to={`/customers/${deal.customer_id}`}>{deal.customer_id}</Link></dd>
        <dt>Автомобиль ID</dt>
        <dd><Link to={`/vehicles/${deal.vehicle_id}`}>{deal.vehicle_id}</Link></dd>
        <dt>Сумма</dt>
        <dd>{deal.amount ? Number(deal.amount).toLocaleString('ru') : '—'}</dd>
        <dt>Этап</dt>
        <dd>{STAGE_LABEL[deal.stage] || deal.stage}</dd>
        {deal.notes && (
          <>
            <dt>Заметки</dt>
            <dd>{deal.notes}</dd>
          </>
        )}
      </dl>
      <p className="customer-view-back">
        <Link to="/deals">← К списку сделок</Link>
      </p>
    </div>
  )
}
