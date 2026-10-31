import type { Schema } from '@/shared/api/types'

export type Guest = Schema<'Guest'>
export type GuestListItem = Schema<'GuestListItem'>
export type MaskedGuestIdentifier = Schema<'MaskedGuestIdentifier'>
export type GuestStatus = Schema<'GuestStatus'>
export type GuestEvent = Schema<'Event'>
export type GuestMovementStep = Schema<'MovementStep'>
export type GuestListSort =
  | 'createdAt'
  | '-createdAt'
  | 'eventCount'
  | '-eventCount'
  | 'totalAmountMinor'
  | '-totalAmountMinor'
  | 'firstSeenAt'
  | '-firstSeenAt'
  | 'lastSeenAt'
  | '-lastSeenAt'
export type GuestEventSort = '-occurredAt' | 'occurredAt' | '-amountMinor' | 'amountMinor'
export type GuestContact = Schema<'ContactLog'>
export type GuestProfile = Schema<'GuestProfile'>
export type GuestIdentifier = Schema<'GuestIdentifier'>
export type GuestTenantSummary = Schema<'GuestTenantSummary'>
export type GuestSegmentSummary = Schema<'GuestSegmentSummary'>
export type IdentifierType = Schema<'IdentifierType'>
export type ContactChannel = Schema<'ExportChannel'>
