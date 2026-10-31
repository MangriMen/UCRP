import type { CategoryRelationType } from './types'

export const DEFAULT_RELATION_WEIGHT_PCT = 100

export const relationTypeLabels: Record<CategoryRelationType, string> = {
  adjacent: 'Смежная',
  competitor: 'Прямой конкурент',
}
