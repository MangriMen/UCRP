import type { Mall } from '../model/types'

export const MALL_WITH_TENANTS_ID = 'mall-0001'

export const mockMalls: Mall[] = [
  {
    id: 'mall-0001',
    code: 'central',
    name: 'Центральный ТЦ',
    isActive: true,
    floors: [1, 2, 3, 4],
  },
  {
    id: 'mall-0002',
    code: 'north',
    name: 'Северный ТЦ',
    isActive: true,
    floors: [1, 2],
  },
]
