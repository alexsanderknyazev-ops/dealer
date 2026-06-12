import { useEffect, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { ArrowLeft, Pencil, Wrench, XCircle } from 'lucide-react'
import * as api from './repairAppointmentsApi'
import { useAuth } from './auth'
import { PageHeader } from '@/components/common/PageHeader'
import { ErrorAlert } from '@/components/common/ErrorAlert'
import { LoadingState } from '@/components/common/LoadingState'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'

const STATUS_LABEL: Record<string, string> = {
  draft: 'Черновик',
  scheduled: 'Запланирована',
  in_progress: 'В работе',
  completed: 'Завершена',
  cancelled: 'Отменена',
}

function formatTs(ts: number) {
  return new Date(ts * 1000).toLocaleString('ru-RU', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

export function RepairAppointmentView() {
  const { id } = useParams()
  const navigate = useNavigate()
  const { logout } = useAuth()
  const [appointment, setAppointment] = useState<api.RepairAppointment | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [acting, setActing] = useState(false)

  function load() {
    if (!id) return
    setLoading(true)
    api
      .getRepairAppointment(id)
      .then(setAppointment)
      .catch(async (e) => {
        if (e instanceof api.ApiError && (e.status === 401 || e.status === 403)) {
          await logout()
          navigate('/login', { replace: true })
          return
        }
        setError(e instanceof Error ? e.message : 'Ошибка загрузки')
      })
      .finally(() => setLoading(false))
  }

  useEffect(() => {
    load()
  }, [id])

  async function handleCreateWorkOrder() {
    if (!id) return
    setActing(true)
    setError(null)
    try {
      const r = await api.createWorkOrderFromAppointment(id)
      navigate(`/work-orders/${r.work_order_id}`)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Не удалось создать заказ-наряд')
    } finally {
      setActing(false)
    }
  }

  async function handleCancel() {
    if (!id || !window.confirm('Отменить запись на ремонт?')) return
    setActing(true)
    setError(null)
    try {
      const a = await api.cancelRepairAppointment(id)
      setAppointment(a)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Не удалось отменить запись')
    } finally {
      setActing(false)
    }
  }

  if (loading) return <LoadingState />
  if (!appointment) return error ? <ErrorAlert message={error} onRetry={load} /> : null

  const canEdit = !appointment.work_order_id && appointment.status !== 'cancelled' && appointment.status !== 'completed'
  const canCreateWO = !appointment.work_order_id && appointment.status !== 'cancelled'
  const canCancel = !appointment.work_order_id && appointment.status !== 'cancelled'

  return (
    <div className="mx-auto w-full max-w-4xl space-y-6">
      <PageHeader
        title={appointment.appointment_number}
        subtitle="Запись на ремонт"
        action={
          <div className="flex flex-wrap gap-2">
            <Button variant="outline" asChild>
              <Link to="/repair-appointments"><ArrowLeft className="mr-1 h-4 w-4" /> К расписанию</Link>
            </Button>
            {canEdit && (
              <Button variant="outline" asChild>
                <Link to={`/repair-appointments/${appointment.id}/edit`}><Pencil className="mr-1 h-4 w-4" /> Редактировать</Link>
              </Button>
            )}
            {canCreateWO && (
              <Button disabled={acting} onClick={handleCreateWorkOrder}>
                <Wrench className="mr-1 h-4 w-4" /> Создать заказ-наряд
              </Button>
            )}
            {canCancel && (
              <Button variant="destructive" disabled={acting} onClick={handleCancel}>
                <XCircle className="mr-1 h-4 w-4" /> Отменить
              </Button>
            )}
          </div>
        }
      />
      {error && <ErrorAlert message={error} onRetry={load} />}

      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <CardTitle className="text-base">Основное</CardTitle>
          <Badge variant="secondary">{STATUS_LABEL[appointment.status] || appointment.status}</Badge>
        </CardHeader>
        <CardContent className="grid gap-3 sm:grid-cols-2 text-sm">
          <div><span className="text-muted-foreground">Время:</span> {formatTs(appointment.scheduled_start)} — {formatTs(appointment.scheduled_end)}</div>
          <div>
            <span className="text-muted-foreground">Клиент:</span>{' '}
            <Link className="text-primary underline" to={`/customers/${appointment.customer_id}`}>
              {appointment.customer_name || appointment.customer_id}
            </Link>
          </div>
          <div>
            <span className="text-muted-foreground">Автомобиль:</span>{' '}
            {appointment.vehicle_id ? (
              <Link className="text-primary underline" to={`/vehicles/${appointment.vehicle_id}`}>
                {appointment.vehicle_label || appointment.vehicle_vin}
              </Link>
            ) : '—'}
          </div>
          {appointment.work_order_id && (
            <div>
              <span className="text-muted-foreground">Заказ-наряд:</span>{' '}
              <Link className="text-primary underline" to={`/work-orders/${appointment.work_order_id}`}>
                {appointment.work_order_number || appointment.work_order_id}
              </Link>
            </div>
          )}
          {appointment.complaint && (
            <div className="sm:col-span-2"><span className="text-muted-foreground">Жалоба:</span> {appointment.complaint}</div>
          )}
          {appointment.notes && (
            <div className="sm:col-span-2"><span className="text-muted-foreground">Примечания:</span> {appointment.notes}</div>
          )}
        </CardContent>
      </Card>

      {appointment.labor?.length > 0 && (
        <Card>
          <CardHeader><CardTitle className="text-base">Работы</CardTitle></CardHeader>
          <CardContent className="p-0">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Код</TableHead>
                  <TableHead>Наименование</TableHead>
                  <TableHead>Кол-во</TableHead>
                  <TableHead>Цена</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {appointment.labor.map((l) => (
                  <TableRow key={l.id}>
                    <TableCell>{l.work_code || '—'}</TableCell>
                    <TableCell>{l.work_name || l.description}</TableCell>
                    <TableCell>{l.quantity}</TableCell>
                    <TableCell>{l.unit_price}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      )}

      {appointment.parts?.length > 0 && (
        <Card>
          <CardHeader><CardTitle className="text-base">Запчасти</CardTitle></CardHeader>
          <CardContent className="p-0">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Артикул</TableHead>
                  <TableHead>Наименование</TableHead>
                  <TableHead>Склад</TableHead>
                  <TableHead>Кол-во</TableHead>
                  <TableHead>Цена</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {appointment.parts.map((p) => (
                  <TableRow key={p.id}>
                    <TableCell>{p.part_sku}</TableCell>
                    <TableCell>{p.part_name}</TableCell>
                    <TableCell>{p.warehouse_name || p.warehouse_id}</TableCell>
                    <TableCell>{p.quantity}</TableCell>
                    <TableCell>{p.unit_price}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      )}
    </div>
  )
}
