import { isSegmentComputing, type SegmentStatus } from '@/entities/segment'
import { Button } from '@/shared/ui/button'
import { ErrorBanner } from '@/shared/ui/error-banner'
import { useRecomputeSegment } from '../api/use-recompute-segment'

interface RecomputeSegmentButtonProps {
  segmentId: string
  status: SegmentStatus
}

export function RecomputeSegmentButton({ segmentId, status }: RecomputeSegmentButtonProps) {
  const mutation = useRecomputeSegment()
  const isBusy = mutation.isPending || isSegmentComputing(status)

  return (
    <div className="flex flex-col gap-2">
      <Button variant="outline" disabled={isBusy} onClick={() => mutation.mutate(segmentId)}>
        {isBusy ? 'Пересчёт…' : 'Пересчитать'}
      </Button>
      {mutation.error && <ErrorBanner error={mutation.error} />}
    </div>
  )
}
