import { StatusBadge } from '@/shared/ui/status-badge'
import { grantStatusLabels, grantStatusTones } from '../model/labels'
import type { GrantStatus } from '../model/types'

export function GrantStatusBadge({ status }: { status: GrantStatus }) {
  return <StatusBadge tone={grantStatusTones[status]}>{grantStatusLabels[status]}</StatusBadge>
}
