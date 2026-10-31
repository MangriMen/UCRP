import { StatusBadge, type StatusTone } from '@/shared/ui/status-badge'
import { mergeResolutionLabels } from '../model/labels'
import type { MergeResolution } from '../model/types'

const toneByResolution: Record<MergeResolution, StatusTone> = {
  pending: 'info',
  confirmed: 'success',
  auto_merged: 'success',
  rejected: 'neutral',
  reverted: 'danger',
}

export function MergeResolutionBadge({ resolution }: { resolution: MergeResolution }) {
  return (
    <StatusBadge tone={toneByResolution[resolution]}>
      {mergeResolutionLabels[resolution]}
    </StatusBadge>
  )
}
