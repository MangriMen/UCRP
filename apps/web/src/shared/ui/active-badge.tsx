import { StatusBadge } from './status-badge'

export function ActiveBadge({ isActive }: { isActive: boolean }) {
  return (
    <StatusBadge tone={isActive ? 'success' : 'neutral'}>
      {isActive ? 'Активен' : 'Неактивен'}
    </StatusBadge>
  )
}
