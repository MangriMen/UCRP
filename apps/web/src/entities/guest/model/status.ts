import type { GuestStatus } from './types'

export const guestStatusLabels: Record<GuestStatus, string> = {
  active: 'Активен',
  merged: 'Объединён',
  deleted: 'Удалён',
}
