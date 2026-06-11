import { Button } from '@/components/ui/button'

export function FormActions({
  submitting,
  submitLabel,
  onCancel,
  disabled,
}: {
  submitting: boolean
  submitLabel: string
  onCancel: () => void
  disabled?: boolean
}) {
  return (
    <div className="flex flex-wrap gap-2 pt-2">
      <Button type="submit" disabled={submitting || disabled}>
        {submitting ? 'Сохранение…' : submitLabel}
      </Button>
      <Button type="button" variant="outline" onClick={onCancel}>
        Отмена
      </Button>
    </div>
  )
}
