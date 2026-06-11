import type { ReactNode } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { LoadingState } from '@/components/common/LoadingState'

export function FormPage({ title, loading, children }: { title: string; loading?: boolean; children: ReactNode }) {
  if (loading) {
    return (
      <div className="mx-auto w-full max-w-2xl">
        <LoadingState />
      </div>
    )
  }

  return (
    <div className="mx-auto w-full max-w-2xl">
      <Card>
        <CardHeader>
          <CardTitle>{title}</CardTitle>
        </CardHeader>
        <CardContent>{children}</CardContent>
      </Card>
    </div>
  )
}
