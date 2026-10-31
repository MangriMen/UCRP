import type { Tenant } from '../model/types'

export const TENANT_WITH_EVENTS_ID = 'tenant-0001'

export const mockTenants: Tenant[] = [
  {
    id: 'tenant-0001',
    code: 'zara',
    name: 'Заря',
    categoryId: 'category-0002',
    categoryName: 'Одежда и обувь',
    mallId: 'mall-0001',
    floor: 1,
    areaSqm: 320,
    priority: 70,
    isActive: true,
    createdAt: '2026-01-15T09:00:00Z',
    updatedAt: '2026-01-15T09:00:00Z',
  },
  {
    id: 'tenant-0002',
    code: 'bistro',
    name: 'Бистро «Уголок»',
    categoryId: 'category-0001',
    categoryName: 'Питание',
    mallId: 'mall-0001',
    floor: 3,
    areaSqm: 85.5,
    priority: 50,
    isActive: true,
    createdAt: '2026-01-15T09:00:00Z',
    updatedAt: '2026-01-15T09:00:00Z',
  },
  {
    id: 'tenant-0003',
    code: 'cinema',
    name: 'Киномир',
    categoryId: 'category-0003',
    categoryName: 'Развлечения',
    mallId: 'mall-0001',
    floor: 4,
    areaSqm: null,
    priority: 40,
    isActive: false,
    createdAt: '2026-01-20T09:00:00Z',
    updatedAt: '2026-03-01T09:00:00Z',
  },
]
