import { useEffect, useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { useAuth } from '@/auth'
import * as api from '@/api'
import { FormPage } from '@/components/common/FormPage'
import { FormField } from '@/components/common/FormField'
import { FormActions } from '@/components/common/FormActions'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { NativeSelect } from '@/components/ui/native-select'
import { Textarea } from '@/components/ui/textarea'

export function ReviewNew() {
  const { getAccessToken } = useAuth()
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const preselectedVehicleId = searchParams.get('vehicle_id') ?? ''
  const [vehicles, setVehicles] = useState<api.ClientVehicle[]>([])
  const [loading, setLoading] = useState(true)
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [vehicleId, setVehicleId] = useState('')
  const [rating, setRating] = useState('5')
  const [text, setText] = useState('')

  useEffect(() => {
    getAccessToken()
      .then((token) => {
        if (!token) throw new Error('Сессия истекла')
        return api.listVehicles(token)
      })
      .then((r) => {
        const items = r.vehicles ?? []
        setVehicles(items)
        const preferred = preselectedVehicleId && items.some((v) => v.vehicle_id === preselectedVehicleId)
          ? preselectedVehicleId
          : items[0]?.vehicle_id ?? ''
        if (preferred) setVehicleId(preferred)
      })
      .catch((e) => setError(e instanceof Error ? e.message : 'Ошибка загрузки'))
      .finally(() => setLoading(false))
  }, [getAccessToken, preselectedVehicleId])

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!vehicleId) {
      setError('Выберите автомобиль')
      return
    }
    setError(null)
    setSubmitting(true)
    try {
      const token = await getAccessToken()
      if (!token) throw new Error('Сессия истекла')
      await api.createReview(token, {
        vehicle_id: vehicleId,
        rating: Number(rating),
        text: text.trim(),
      })
      navigate('/reviews', { replace: true })
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Ошибка сохранения')
      setSubmitting(false)
    }
  }

  return (
    <FormPage title="Новый отзыв" loading={loading}>
      <form onSubmit={handleSubmit} className="space-y-4">
        {error && (
          <Alert variant="destructive">
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}
        {vehicles.length === 0 && !loading ? (
          <Alert>
            <AlertDescription>Сначала добавьте автомобиль в разделе «Мои авто».</AlertDescription>
          </Alert>
        ) : (
          <>
            <FormField label="Автомобиль" htmlFor="vehicle_id" required>
              <NativeSelect id="vehicle_id" value={vehicleId} onChange={(e) => setVehicleId(e.target.value)} required>
                {vehicles.map((v) => (
                  <option key={v.id} value={v.vehicle_id}>
                    {v.make} {v.model} ({v.vin})
                  </option>
                ))}
              </NativeSelect>
            </FormField>
            <FormField label="Оценка" htmlFor="rating" required>
              <NativeSelect id="rating" value={rating} onChange={(e) => setRating(e.target.value)}>
                <option value="5">5 — Отлично</option>
                <option value="4">4 — Хорошо</option>
                <option value="3">3 — Нормально</option>
                <option value="2">2 — Плохо</option>
                <option value="1">1 — Очень плохо</option>
              </NativeSelect>
            </FormField>
            <FormField label="Отзыв" htmlFor="text" required>
              <Textarea id="text" value={text} onChange={(e) => setText(e.target.value)} rows={5} required placeholder="Расскажите о визите в дилерский центр…" />
            </FormField>
            <FormActions submitting={submitting} submitLabel="Отправить" onCancel={() => navigate('/reviews')} disabled={vehicles.length === 0} />
          </>
        )}
      </form>
    </FormPage>
  )
}
