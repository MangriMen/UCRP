import type { TenantCategory } from '../model/types'

export const CATEGORY_IN_USE_ID = 'category-0001'

export const mockTenantCategories: TenantCategory[] = [
  {
    id: 'category-0001',
    code: 'food',
    name: 'Питание',
    description: 'Рестораны, кафе и фудкорт',
    createdAt: '2026-01-10T09:00:00Z',
    updatedAt: '2026-01-10T09:00:00Z',
  },
  {
    id: 'category-0002',
    code: 'fashion',
    name: 'Одежда и обувь',
    description: null,
    createdAt: '2026-01-10T09:00:00Z',
    updatedAt: '2026-01-10T09:00:00Z',
  },
  {
    id: 'category-0003',
    code: 'entertainment',
    name: 'Развлечения',
    description: 'Кинотеатр, детские зоны',
    createdAt: '2026-01-10T09:00:00Z',
    updatedAt: '2026-02-01T09:00:00Z',
  },
]
