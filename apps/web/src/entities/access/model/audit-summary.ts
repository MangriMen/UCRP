import { grantStatusLabels } from './labels'
import type { AccessAuditEntry, GrantStatus } from './types'

const field = (value: unknown, key: string): unknown =>
  typeof value === 'object' && value !== null ? (value as Record<string, unknown>)[key] : undefined

const text = (value: unknown) => (typeof value === 'string' ? value : '')

const statusLabel = (value: unknown) => {
  const status = text(value)
  return status in grantStatusLabels ? grantStatusLabels[status as GrantStatus] : status
}

export const isGrantAuditAction = (action: string) => action.startsWith('access.grant.')

export function auditSummary(entry: AccessAuditEntry): string {
  const { details } = entry
  if (typeof details === 'string') return details
  if (entry.action === 'access.denied') {
    return [
      text(field(details, 'method')),
      text(field(details, 'route')),
      text(field(details, 'code')),
    ]
      .filter(Boolean)
      .join(' · ')
  }
  if (entry.action === 'access.simulate') {
    return `Субъект: ${text(field(details, 'subjectType'))} ${text(field(details, 'subjectId'))}`
  }
  const before = field(details, 'before')
  const after = field(details, 'after')
  const parts: string[] = []
  const from = statusLabel(field(before, 'Status'))
  const to = statusLabel(field(after, 'Status'))
  if (to && from !== to) parts.push(from ? `${from} → ${to}` : to)
  const reason = text(field(after, 'Reason'))
  if (reason) parts.push(`Причина: ${reason}`)
  return parts.join(' · ')
}
