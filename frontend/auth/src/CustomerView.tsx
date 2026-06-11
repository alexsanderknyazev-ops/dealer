import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import type { Customer } from './customersApi'
import * as api from './customersApi'
import { EntityViewPage } from '@/components/common/EntityViewPage'

export function CustomerView() {
  const { id } = useParams()
  const navigate = useNavigate()
  const [customer, setCustomer] = useState<Customer | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!id) return
    api
      .getCustomer(id)
      .then(setCustomer)
      .catch((e) => setError(e instanceof Error ? e.message : 'Ошибка загрузки'))
      .finally(() => setLoading(false))
  }, [id])

  async function handleDelete() {
    if (!id || !customer || !confirm(`Удалить клиента ${customer.name}?`)) return
    try {
      await api.deleteCustomer(id)
      navigate('/customers', { replace: true })
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Ошибка удаления')
    }
  }

  return (
    <EntityViewPage
      title={customer?.name ?? 'Клиент'}
      backTo="/customers"
      backLabel="К списку клиентов"
      editTo={id ? `/customers/${id}/edit` : undefined}
      onDelete={customer ? handleDelete : undefined}
      loading={loading}
      error={error || (!customer && !loading ? 'Не найден' : null)}
      fields={
        customer
          ? [
              { label: 'Email', value: customer.email || '—' },
              { label: 'Телефон', value: customer.phone || '—' },
              { label: 'Тип', value: customer.customer_type === 'legal' ? 'Юр. лицо' : 'Физ. лицо' },
              ...(customer.inn ? [{ label: 'ИНН', value: customer.inn }] : []),
              ...(customer.address ? [{ label: 'Адрес', value: customer.address }] : []),
              ...(customer.notes ? [{ label: 'Заметки', value: customer.notes }] : []),
            ]
          : []
      }
    />
  )
}
