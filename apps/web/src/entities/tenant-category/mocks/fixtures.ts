import type { TenantCategory, TenantCategoryRelation } from '../model/types'

export const CATEGORY_IN_USE_ID = 'category-0001'

export const mockTenantCategories: TenantCategory[] = [
  {
    id: 'category-0001',
    code: 'food',
    name: 'Питание',
    description: 'Рестораны, кафе и фудкорт',
    isActive: true,
    createdAt: '2026-01-10T09:00:00Z',
    updatedAt: '2026-01-10T09:00:00Z',
  },
  {
    id: 'category-0002',
    code: 'fashion',
    name: 'Одежда и обувь',
    description: null,
    isActive: true,
    createdAt: '2026-01-10T09:00:00Z',
    updatedAt: '2026-01-10T09:00:00Z',
  },
  {
    id: 'category-0003',
    code: 'entertainment',
    name: 'Развлечения',
    description: 'Кинотеатр, детские зоны',
    isActive: true,
    createdAt: '2026-01-10T09:00:00Z',
    updatedAt: '2026-02-01T09:00:00Z',
  },
  {
    id: 'category-0004',
    code: 'seasonal',
    name: 'Сезонные товары',
    description: null,
    isActive: false,
    createdAt: '2026-01-10T09:00:00Z',
    updatedAt: '2026-01-10T09:00:00Z',
  },
]

export const mockTenantCategoryRelations: TenantCategoryRelation[] = [
  {
    categoryId: 'category-0003',
    relatedCategoryId: 'category-0001',
    relationType: 'adjacent',
    weightPct: 70,
  },
]
