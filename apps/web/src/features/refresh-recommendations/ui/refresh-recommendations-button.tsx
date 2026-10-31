import { Button } from '@/shared/ui/button'
import { ErrorBanner } from '@/shared/ui/error-banner'
import { useRefreshRecommendations } from '../api/use-refresh-recommendations'

export function RefreshRecommendationsButton({ guestId }: { guestId: string }) {
  const mutation = useRefreshRecommendations()

  return (
    <div className="flex flex-col gap-2">
      <Button
        variant="outline"
        size="sm"
        disabled={mutation.isPending}
        onClick={() => mutation.mutate(guestId)}
      >
        {mutation.isPending ? 'Обновление…' : 'Обновить'}
      </Button>
      {mutation.error && <ErrorBanner error={mutation.error} />}
    </div>
  )
}
