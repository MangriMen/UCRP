import type { Project } from '../model/types'

export const PROJECT_WITH_EVENTS_ID = 'project-0001'

export const mockProjects: Project[] = [
  {
    id: 'project-0001',
    tenantId: 'tenant-0001',
    tenantName: 'Заря',
    code: 'zara-main',
    name: 'Основной зал',
    isActive: true,
    createdAt: '2026-01-16T09:00:00Z',
    updatedAt: '2026-01-16T09:00:00Z',
  },
  {
    id: 'project-0002',
    tenantId: 'tenant-0002',
    tenantName: 'Бистро «Уголок»',
    code: 'bistro-delivery',
    name: 'Доставка',
    isActive: true,
    createdAt: '2026-02-02T09:00:00Z',
    updatedAt: '2026-02-02T09:00:00Z',
  },
]
