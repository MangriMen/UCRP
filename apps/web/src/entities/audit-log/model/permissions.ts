import type { PermissionCode } from '@/entities/user'

export const AUDIT_LOG_PERMISSIONS: PermissionCode[] = [
  'audit.read',
  'field.guest.phone.read',
  'field.guest.email.read',
  'field.guest.card.read',
  'field.event.raw_payload.read',
]
