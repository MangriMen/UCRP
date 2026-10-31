import type { MergeResolution } from './types'

export const mergeResolutionLabels: Record<MergeResolution, string> = {
  pending: 'Ожидает решения',
  confirmed: 'Объединены',
  rejected: 'Отклонено',
  auto_merged: 'Объединены автоматически',
  reverted: 'Разъединены',
}

export const mergeConflictMessages: Record<string, string> = {
  already_resolved: 'Пара уже обработана. Обновите данные и проверьте текущий статус.',
  merge_blocked:
    'Для этой пары действует запрет объединения. Снимите запрет в разделе «Запреты объединения» и повторите.',
  conflict: 'Состояние профилей изменилось. Обновите данные и повторите попытку.',
}
