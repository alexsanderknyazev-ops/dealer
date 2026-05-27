import { useEffect, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import type { Customer } from './customersApi'
import * as api from './customersApi'
import './CustomerView.css'

export function CustomerView() {
  const { id } = useParams()
  const navigate = useNavigate()
  const [customer, setCustomer] = useState<Customer | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!id) return
    api.getCustomer(id)
      .then(setCustomer)
      .catch((e) => setError(e.message))
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

  if (loading) return <div className="main loading">Загрузка…</div>
  if (error || !customer) return <div className="form-error">{error || 'Не найден'}</div>

  return (
    <div className="customer-view">
      <div className="customer-view-header">
        <h1 className="customer-view-name">{customer.name}</h1>
        <div className="customer-view-actions">
          <Link to={`/customers/${id}/edit`} className="customer-view-btn customer-view-edit">Редактировать</Link>
          <button type="button" onClick={handleDelete} className="customer-view-btn customer-view-delete">Удалить</button>
        </div>
      </div>
      <dl className="customer-view-dl">
        <dt>Email</dt>
        <dd>{customer.email || '—'}</dd>
        <dt>Телефон</dt>
        <dd>{customer.phone || '—'}</dd>
        <dt>Тип</dt>
        <dd>{customer.customer_type === 'legal' ? 'Юр. лицо' : 'Физ. лицо'}</dd>
        {customer.inn && (
          <>
            <dt>ИНН</dt>
            <dd>{customer.inn}</dd>
          </>
        )}
        {customer.address && (
          <>
            <dt>Адрес</dt>
            <dd>{customer.address}</dd>
          </>
        )}
        {customer.notes && (
          <>
            <dt>Заметки</dt>
            <dd>{customer.notes}</dd>
          </>
        )}
      </dl>
      <p className="customer-view-back">
        <Link to="/customers">← К списку клиентов</Link>
      </p>
    </div>
  )
}
