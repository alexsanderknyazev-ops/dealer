import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { Car, MessageSquare, Star } from 'lucide-react'
import { useAuth } from '@/auth'
import * as api from '@/api'
import { PageHeader } from '@/components/common/PageHeader'
import { LoadingState } from '@/components/common/LoadingState'
import { ErrorAlert } from '@/components/common/ErrorAlert'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Alert, AlertDescription } from '@/components/ui/alert'

function invitationLabel(inv: api.ReviewInvitation): string {
  if (inv.service_kind === 'sale') return 'Оцените покупку автомобиля'
  return 'Оцените обслуживание в сервисе'
}

export function Dashboard() {
  const { getAccessToken, user } = useAuth()
  const [profile, setProfile] = useState<api.ClientProfile | null>(null)
  const [invitations, setInvitations] = useState<api.ReviewInvitation[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [dismissingId, setDismissingId] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    getAccessToken()
      .then(async (token) => {
        if (!token) throw new Error('Сессия истекла')
        const [p, inv] = await Promise.all([api.getProfile(token), api.listReviewInvitations(token)])
        return { p, inv }
      })
      .then(({ p, inv }) => {
        if (!cancelled) {
          setProfile(p)
          setInvitations(inv.invitations ?? [])
        }
      })
      .catch((e) => {
        if (!cancelled) setError(e instanceof Error ? e.message : 'Ошибка загрузки')
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [getAccessToken])

  async function handleDismiss(id: string) {
    setDismissingId(id)
    try {
      const token = await getAccessToken()
      if (!token) throw new Error('Сессия истекла')
      await api.dismissReviewInvitation(token, id)
      setInvitations((prev) => prev.filter((i) => i.id !== id))
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Не удалось скрыть предложение')
    } finally {
      setDismissingId(null)
    }
  }

  if (loading) return <LoadingState />
  if (error) return <ErrorAlert message={error} />

  return (
    <div className="mx-auto w-full max-w-2xl space-y-6">
      <PageHeader title="Профиль" subtitle={<span className="text-muted-foreground">{user?.email}</span>} />

      {invitations.length > 0 && (
        <div className="space-y-3">
          {invitations.map((inv) => (
            <Alert key={inv.id}>
              <Star className="h-4 w-4" />
              <AlertDescription className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                <span>{invitationLabel(inv)}</span>
                <span className="flex shrink-0 gap-2">
                  <Button size="sm" asChild>
                    <Link to={`/reviews/new?vehicle_id=${inv.vehicle_id}`}>Оставить отзыв</Link>
                  </Button>
                  <Button
                    size="sm"
                    variant="ghost"
                    disabled={dismissingId === inv.id}
                    onClick={() => handleDismiss(inv.id)}
                  >
                    Позже
                  </Button>
                </span>
              </AlertDescription>
            </Alert>
          ))}
        </div>
      )}

      <Card>
        <CardHeader>
          <CardTitle>{profile?.full_name || 'Клиент'}</CardTitle>
          <CardDescription>
            {profile?.phone || 'Телефон не указан'} · {profile?.vehicles?.length ?? 0} авто
          </CardDescription>
        </CardHeader>
        <CardContent className="grid gap-2 sm:grid-cols-2">
          <Button variant="outline" className="justify-start" asChild>
            <Link to="/vehicles">
              <Car className="mr-2 h-4 w-4" />
              Мои автомобили
            </Link>
          </Button>
          <Button variant="outline" className="justify-start" asChild>
            <Link to="/reviews">
              <MessageSquare className="mr-2 h-4 w-4" />
              Мои отзывы
            </Link>
          </Button>
        </CardContent>
      </Card>
    </div>
  )
}
