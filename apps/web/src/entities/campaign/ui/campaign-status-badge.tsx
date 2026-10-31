import { StatusBadge, type StatusTone } from '@/shared/ui/status-badge'
import { campaignStatusLabels } from '../model/status'
import type { CampaignStatus } from '../model/types'

const toneByStatus: Record<CampaignStatus, StatusTone> = {
  draft: 'neutral',
  active: 'info',
  finished: 'success',
  archived: 'neutral',
}

export function CampaignStatusBadge({ status }: { status: CampaignStatus }) {
  return <StatusBadge tone={toneByStatus[status]}>{campaignStatusLabels[status]}</StatusBadge>
}
