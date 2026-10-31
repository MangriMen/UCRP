import type { Recommendation, SegmentRecommendation } from '../model/types'

export const mockGuestRecommendations: Record<string, Recommendation[]> = {
  'guest-0001': [
    {
      id: 'rec-0001',
      guestId: 'guest-0001',
      tenantId: 'tenant-0003',
      tenantName: 'Кинотеатр «Синема»',
      categoryName: 'Развлечения',
      floor: 3,
      rank: 1,
      score: 0.78,
      explanation: 'Похожие гости посещают кинотеатр; спорт и развлечения — смежные категории',
      algorithmVersion: 'v1',
      factors: [
        { factorCode: 'similar_guests', rawValue: 0.8, weight: 0.4, contribution: 0.32 },
        { factorCode: 'category_adjacency', rawValue: 0.7, weight: 0.3, contribution: 0.21 },
        { factorCode: 'distance', rawValue: 0.5, weight: 0.2, contribution: 0.1 },
      ],
      computedAt: '2026-10-30T06:00:00Z',
    },
    {
      id: 'rec-0002',
      guestId: 'guest-0001',
      tenantId: 'tenant-0004',
      tenantName: 'Кофейня «Зерно»',
      categoryName: 'Кафе',
      floor: 1,
      rank: 2,
      score: 0.64,
      explanation: 'Гость часто заходит на первый этаж',
      algorithmVersion: 'v1',
      factors: [{ factorCode: 'recency', rawValue: 0.9, weight: 0.3, contribution: 0.27 }],
      computedAt: '2026-10-30T06:00:00Z',
    },
    {
      id: 'rec-0003',
      guestId: 'guest-0001',
      tenantId: 'tenant-0005',
      tenantName: 'Спортмастер',
      categoryName: 'Спорт',
      floor: 2,
      rank: 3,
      score: 0.51,
      explanation: 'Приоритетный арендатор в смежной категории',
      algorithmVersion: 'v1',
      computedAt: '2026-10-30T06:00:00Z',
    },
  ],
}

export const mockSegmentRecommendations: SegmentRecommendation[] = [
  {
    tenantId: 'tenant-0003',
    tenantName: 'Кинотеатр «Синема»',
    categoryId: 'category-0003',
    guests: 64,
    averageScore: 0.71,
    suppressed: false,
  },
  {
    tenantId: 'tenant-0004',
    tenantName: 'Кофейня «Зерно»',
    categoryId: 'category-0001',
    guests: 41,
    averageScore: 0.62,
    suppressed: false,
  },
  {
    tenantId: 'tenant-0005',
    tenantName: 'Спортмастер',
    categoryId: null,
    guests: 22,
    averageScore: 0.55,
    suppressed: false,
  },
]
