import type { AccessGrant, GrantStatus } from './types'

export type GrantTransition = 'activate' | 'suspend' | 'revoke'

const transitionSources: Record<GrantTransition, GrantStatus[]> = {
  activate: ['draft', 'suspended'],
  suspend: ['active'],
  revoke: ['draft', 'active', 'suspended'],
}

const TRANSITIONS: GrantTransition[] = ['activate', 'suspend', 'revoke']

export const isGrantEnded = (grant: AccessGrant, now = new Date()) =>
  grant.validTo !== null && grant.validTo !== undefined && new Date(grant.validTo) <= now

export const isGrantEditable = (grant: AccessGrant, now = new Date()) =>
  grant.status !== 'revoked' && grant.status !== 'expired' && !isGrantEnded(grant, now)

export function availableTransitions(grant: AccessGrant, now = new Date()): GrantTransition[] {
  if (isGrantEnded(grant, now)) return []
  return TRANSITIONS.filter((transition) => transitionSources[transition].includes(grant.status))
}
