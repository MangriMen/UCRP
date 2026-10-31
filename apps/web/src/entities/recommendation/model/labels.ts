import type { RecommendationFactor } from './types'

export const factorLabels: Record<RecommendationFactor['factorCode'], string> = {
  similar_guests: 'Похожие гости',
  category_adjacency: 'Смежные категории',
  recency: 'Давность визитов',
  distance: 'Расстояние',
  tenant_priority: 'Приоритет арендатора',
}
