import type { MergeBlockReason } from './types'

export const mergeBlockReasonLabels: Record<MergeBlockReason, string> = {
  manual_unmerge: 'Ручное разъединение',
  rejected: 'Отклонённое объединение',
}
