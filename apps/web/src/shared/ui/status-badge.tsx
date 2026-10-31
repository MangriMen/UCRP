import type { ComponentProps } from 'react'
import { Badge } from './badge'

export type StatusTone = 'neutral' | 'info' | 'success' | 'danger'

const variantByTone: Record<StatusTone, ComponentProps<typeof Badge>['variant']> = {
  neutral: 'outline',
  info: 'secondary',
  success: 'default',
  danger: 'destructive',
}

interface StatusBadgeProps {
  tone: StatusTone
  children: string
}

export function StatusBadge({ tone, children }: StatusBadgeProps) {
  return <Badge variant={variantByTone[tone]}>{children}</Badge>
}
