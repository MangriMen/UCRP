import type { Me, User } from '../model/types'

export const MOCK_PASSWORD = 'password'

export const mockUsers: Me[] = [
  {
    userId: 'user-marketer',
    email: 'marketer@ucrp.test',
    displayName: 'Мария Маркетолог',
    role: 'marketer',
    policyVersion: 1,
    tenantId: null,
    permissions: ['guests:read', 'segments:write', 'exports:write', 'imports:write'],
  },
  {
    userId: 'user-manager',
    email: 'manager@ucrp.test',
    displayName: 'Максим Менеджер',
    role: 'manager',
    policyVersion: 1,
    tenantId: null,
    permissions: ['guests:read', 'guests:merge', 'imports:write', 'admin:write'],
  },
  {
    userId: 'user-tenant',
    email: 'tenant@ucrp.test',
    displayName: 'Тимур Арендатор',
    role: 'tenant',
    policyVersion: 1,
    tenantId: 'tenant-0001',
    permissions: ['dashboard:read', 'campaigns:write'],
  },
  {
    userId: 'user-analyst',
    email: 'analyst@ucrp.test',
    displayName: 'Анна Аналитик',
    role: 'analyst',
    policyVersion: 1,
    tenantId: null,
    permissions: ['guests:read', 'dashboard:read', 'dq:read'],
  },
]

export const SELF_USER_ID = 'user-manager'

export const mockUserRecords: User[] = [
  ...mockUsers.map((u) => ({
    id: u.userId,
    email: u.email,
    displayName: u.displayName,
    role: u.role,
    tenantId: u.tenantId ?? null,
    isActive: true,
    lastLoginAt: '2026-10-30T08:00:00Z',
    createdAt: '2026-01-05T09:00:00Z',
    updatedAt: '2026-01-05T09:00:00Z',
  })),
  {
    id: 'user-marketer-2',
    email: 'olga@ucrp.test',
    displayName: 'Ольга Маркетолог',
    role: 'marketer',
    tenantId: null,
    isActive: false,
    lastLoginAt: null,
    createdAt: '2026-03-12T09:00:00Z',
    updatedAt: '2026-06-01T09:00:00Z',
  },
]
