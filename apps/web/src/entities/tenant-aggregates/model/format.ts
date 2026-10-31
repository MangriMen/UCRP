import { formatMoney, formatNumber } from '@/shared/lib/format'

export const formatAggregateCell = (value: number | null | undefined, threshold: number) =>
  value === null || value === undefined ? `меньше ${threshold}` : formatNumber(value)

export const formatAggregateMoney = (value: number | null | undefined, threshold: number) =>
  value === null || value === undefined ? `меньше ${threshold} гостей` : formatMoney(value)
