import { useEffect, useMemo, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { Calendar, Plus } from 'lucide-react'
import * as api from './repairAppointmentsApi'
import { useAuth } from './auth'
import { PageHeader } from '@/components/common/PageHeader'
import { ErrorAlert } from '@/components/common/ErrorAlert'
import { LoadingState } from '@/components/common/LoadingState'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'

const STATUS_LABEL: Record<string, string> = {
  draft: 'Черновик',
  scheduled: 'Запланирована',
  in_progress: 'В работе',
  completed: 'Завершена',
  cancelled: 'Отменена',
}

function todayISO() {
  return new Date().toISOString().slice(0, 10)
}

function dayBounds(dateStr: string) {
  const start = new Date(`${dateStr}T00:00:00Z`).getTime() / 1000
  return { from: start, to: start + 86400 }
}

export function RepairAppointments() {
  const navigate = useNavigate()
  const { logout } = useAuth()
  const [date, setDate] = useState(todayISO())
  const [slots, setSlots] = useState<api.RepairAppointmentSlot[]>([])
  const [appointments, setAppointments] = useState<api.RepairAppointment[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const bounds = useMemo(() => dayBounds(date), [date])

  function load() {
    setLoading(true)
    setError(null)
    Promise.all([
      api.listSlots(date),
      api.listRepairAppointments({ limit: 50, date_from: bounds.from, date_to: bounds.to }),
    ])
      .then(([s, list]) => {
        setSlots(s.slots ?? [])
        setAppointments(list.appointments ?? [])
      })
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
  }, [date])

  function openSlot(slot: api.RepairAppointmentSlot) {
    if (!slot.available) return
    const sp = new URLSearchParams({
      start: String(slot.start_at),
      end: String(slot.end_at),
      date,
    })
    navigate(`/repair-appointments/new?${sp}`)
  }

  if (loading) return <LoadingState />

  return (
    <div className="mx-auto w-full max-w-6xl space-y-6">
      <PageHeader
        title="Запись на ремонт"
        subtitle="Выберите свободный слот с 08:00 до 13:00"
        action={
          <Button variant="outline" asChild>
            <Link to={`/repair-appointments/new?date=${date}`}>
              <Plus className="mr-1 h-4 w-4" /> Новая запись
            </Link>
          </Button>
        }
      />
      {error && <ErrorAlert message={error} onRetry={load} />}

      <Card>
        <CardHeader className="flex flex-row items-center justify-between gap-4">
          <CardTitle className="flex items-center gap-2 text-base">
            <Calendar className="h-4 w-4" /> Расписание
          </CardTitle>
          <Input type="date" className="w-auto" value={date} onChange={(e) => setDate(e.target.value)} />
        </CardHeader>
        <CardContent>
          <div className="grid gap-2 sm:grid-cols-2 lg:grid-cols-3">
            {slots.map((slot) => (
              <Button
                key={slot.start_at}
                variant={slot.available ? 'outline' : 'ghost'}
                disabled={!slot.available}
                className="h-auto flex-col items-start py-3"
                onClick={() => openSlot(slot)}
              >
                <span className="font-medium">{slot.label}</span>
                <span className="text-xs text-muted-foreground">
                  {slot.available ? 'Свободно' : 'Занято'}
                </span>
              </Button>
            ))}
          </div>
          {slots.length === 0 && <p className="text-sm text-muted-foreground">Нет слотов на выбранную дату</p>}
        </CardContent>
      </Card>

      <Card>
        <CardHeader><CardTitle className="text-base">Записи на {date}</CardTitle></CardHeader>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Номер</TableHead>
                <TableHead>Время</TableHead>
                <TableHead>Клиент</TableHead>
                <TableHead>Автомобиль</TableHead>
                <TableHead>Статус</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {appointments.map((a) => (
                <TableRow key={a.id}>
                  <TableCell>
                    <Link className="text-primary underline" to={`/repair-appointments/${a.id}`}>
                      {a.appointment_number}
                    </Link>
                  </TableCell>
                  <TableCell>
                    {new Date(a.scheduled_start * 1000).toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' })}
                    –
                    {new Date(a.scheduled_end * 1000).toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' })}
                  </TableCell>
                  <TableCell>{a.customer_name || a.customer_id}</TableCell>
                  <TableCell>{a.vehicle_label || a.vehicle_vin || '—'}</TableCell>
                  <TableCell>
                    <Badge variant="secondary">{STATUS_LABEL[a.status] || a.status}</Badge>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
          {appointments.length === 0 && (
            <p className="p-4 text-sm text-muted-foreground">На эту дату записей нет</p>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
