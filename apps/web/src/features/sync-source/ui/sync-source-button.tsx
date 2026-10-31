import { Button } from '@/shared/ui/button'
import { ErrorBanner } from '@/shared/ui/error-banner'
import { useSyncSource } from '../api/use-sync-source'

interface SyncSourceButtonProps {
  sourceId: string
  isRunning: boolean
}

export function SyncSourceButton({ sourceId, isRunning }: SyncSourceButtonProps) {
  const mutation = useSyncSource()
  const isBusy = mutation.isPending || isRunning

  return (
    <div className="flex flex-col items-end gap-2">
      <Button
        variant="outline"
        size="sm"
        disabled={isBusy}
        onClick={() => mutation.mutate(sourceId)}
      >
        {isBusy ? 'Синхронизация…' : 'Синхронизировать'}
      </Button>
      {mutation.error && <ErrorBanner error={mutation.error} />}
    </div>
  )
}
