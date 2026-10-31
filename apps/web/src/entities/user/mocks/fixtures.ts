import type { Me } from '../model/types'

export const MOCK_PASSWORD = 'password'

export const mockUsers: Me[] = [
  {
    userId: 'user-marketer',
    email: 'marketer@ucrp.test',
    displayName: 'Мария Маркетолог',
    role: 'marketer',
    tenantId: null,
    permissions: ['guests:read', 'segments:write', 'exports:write', 'imports:write'],
  },
  {
    userId: 'user-manager',
    email: 'manager@ucrp.test',
    displayName: 'Максим Менеджер',
    role: 'manager',
    tenantId: null,
    permissions: ['guests:read', 'guests:merge', 'imports:write', 'admin:write'],
  },
  {
    userId: 'user-tenant',
    email: 'tenant@ucrp.test',
    displayName: 'Тимур Арендатор',
    role: 'tenant',
    tenantId: 'tenant-0001',
    permissions: ['dashboard:read', 'campaigns:write'],
  },
  {
    userId: 'user-analyst',
    email: 'analyst@ucrp.test',
    displayName: 'Анна Аналитик',
    role: 'analyst',
    tenantId: null,
    permissions: ['guests:read', 'dashboard:read', 'dq:read'],
  },
]
