import type { SegmentStatus } from './types'

export const segmentStatusLabels: Record<SegmentStatus, string> = {
  pending: 'Ожидает расчёта',
  computing: 'Считается',
  ready: 'Готов',
  failed: 'Ошибка расчёта',
  archived: 'В архиве',
}

export const isSegmentComputing = (status: SegmentStatus) =>
  status === 'pending' || status === 'computing'
