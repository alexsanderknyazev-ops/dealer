import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { Car, MessageSquare } from 'lucide-react'
import { useAuth } from '@/auth'
import * as api from '@/api'
import { PageHeader } from '@/components/common/PageHeader'
import { LoadingState } from '@/components/common/LoadingState'
import { ErrorAlert } from '@/components/common/ErrorAlert'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'

export function Dashboard() {
  const { getAccessToken, user } = useAuth()
  const [profile, setProfile] = useState<api.ClientProfile | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    getAccessToken()
      .then((token) => {
        if (!token) throw new Error('Сессия истекла')
        return api.getProfile(token)
      })
      .then((p) => {
        if (!cancelled) setProfile(p)
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

  if (loading) return <LoadingState />
  if (error) return <ErrorAlert message={error} />

  return (
    <div className="mx-auto w-full max-w-2xl space-y-6">
      <PageHeader title="Профиль" subtitle={<span className="text-muted-foreground">{user?.email}</span>} />

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
