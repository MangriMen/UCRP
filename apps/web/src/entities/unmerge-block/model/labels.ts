import type { UnmergeBlockReason } from './types'

export const unmergeBlockReasonLabels: Record<UnmergeBlockReason, string> = {
  manual_unmerge: 'Ручное разъединение',
  manual_reject: 'Отклонённая склейка',
  auto_conflict: 'Конфликт при автосклейке',
}
