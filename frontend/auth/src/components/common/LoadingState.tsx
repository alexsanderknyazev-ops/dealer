import { Skeleton } from '@/components/ui/skeleton'

export function LoadingState({ label = 'Загрузка…' }: { label?: string }) {
  return (
    <div className="space-y-3" aria-busy="true" aria-label={label}>
      <p className="text-sm text-muted-foreground">{label}</p>
      <Skeleton className="h-10 w-full" />
      <Skeleton className="h-10 w-full" />
      <Skeleton className="h-10 w-3/4" />
    </div>
  )
}
