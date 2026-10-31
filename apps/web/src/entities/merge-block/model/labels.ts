import type { MergeBlockReason } from './types'

export const mergeBlockReasonLabels: Record<MergeBlockReason, string> = {
  manual_unmerge: 'Ручное разъединение',
  manual_reject: 'Отклонённое объединение',
  auto_conflict: 'Конфликт при автоматическом объединении',
}
