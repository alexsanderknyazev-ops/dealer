import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import type { Employee } from './employeesApi'
import * as api from './employeesApi'
import { EntityViewPage } from '@/components/common/EntityViewPage'
import { Badge } from '@/components/ui/badge'

function fmtTime(ts: number) {
  return ts ? new Date(ts * 1000).toLocaleString('ru-RU') : '—'
}

export function EmployeeView() {
  const { id } = useParams()
  const navigate = useNavigate()
  const [employee, setEmployee] = useState<Employee | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!id) return
    api
      .getEmployee(id)
      .then(setEmployee)
      .catch((e) => setError(e instanceof Error ? e.message : 'Ошибка загрузки'))
      .finally(() => setLoading(false))
  }, [id])

  async function handleDelete() {
    if (!id || !employee || !confirm(`Удалить сотрудника ${employee.full_name}?`)) return
    try {
      await api.deleteEmployee(id)
      navigate('/employees', { replace: true })
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Ошибка удаления')
    }
  }

  return (
    <EntityViewPage
      title={employee?.full_name ?? 'Сотрудник'}
      backTo="/employees"
      backLabel="К списку сотрудников"
      editTo={id ? `/employees/${id}/edit` : undefined}
      onDelete={employee ? handleDelete : undefined}
      loading={loading}
      error={error || (!employee && !loading ? 'Не найден' : null)}
      fields={
        employee
          ? [
              { label: 'Должность', value: api.positionLabel(employee.position) },
              { label: 'Отдел', value: employee.department || '—' },
              { label: 'Телефон', value: employee.phone || '—' },
              {
                label: 'Статус',
                value: (
                  <Badge variant={employee.active ? 'default' : 'secondary'}>
                    {employee.active ? 'Активен' : 'Неактивен'}
                  </Badge>
                ),
              },
              ...(employee.user_id
                ? [{ label: 'User ID', value: <span className="font-mono text-xs">{employee.user_id}</span> }]
                : []),
              { label: 'Создан', value: fmtTime(employee.created_at) },
              { label: 'Обновлён', value: fmtTime(employee.updated_at) },
            ]
          : []
      }
    />
  )
}
