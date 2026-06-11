import { useEffect, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import type { Deal } from './dealsApi'
import * as api from './dealsApi'
import { EntityViewPage } from '@/components/common/EntityViewPage'

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
    api
      .getDeal(id)
      .then(setDeal)
      .catch((e) => setError(e instanceof Error ? e.message : 'Ошибка загрузки'))
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

  if (!deal && !loading && !error) {
    return <EntityViewPage title="Сделка" backTo="/deals" backLabel="К списку сделок" fields={[]} error="Не найдено" />
  }

  return (
    <EntityViewPage
      title={deal ? `Сделка ${deal.id.slice(0, 8)}…` : 'Сделка'}
      backTo="/deals"
      backLabel="К списку сделок"
      editTo={id ? `/deals/${id}/edit` : undefined}
      onDelete={deal ? handleDelete : undefined}
      loading={loading}
      error={error}
      fields={
        deal
          ? [
              {
                label: 'Клиент ID',
                value: (
                  <Link to={`/customers/${deal.customer_id}`} className="text-primary hover:underline">
                    {deal.customer_id}
                  </Link>
                ),
              },
              {
                label: 'Автомобиль ID',
                value: (
                  <Link to={`/vehicles/${deal.vehicle_id}`} className="text-primary hover:underline">
                    {deal.vehicle_id}
                  </Link>
                ),
              },
              { label: 'Сумма', value: deal.amount ? Number(deal.amount).toLocaleString('ru') : '—' },
              { label: 'Этап', value: STAGE_LABEL[deal.stage] || deal.stage },
              ...(deal.notes ? [{ label: 'Заметки', value: deal.notes }] : []),
            ]
          : []
      }
    />
  )
}
