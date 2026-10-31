import type { Role } from './types'

export interface UserListFilters {
  q?: string
  role?: Role
  page?: number
  pageSize?: number
}

export const userKeys = {
  me: ['me'] as const,
  all: ['users'] as const,
  lists: () => [...userKeys.all, 'list'] as const,
  list: (filters: UserListFilters) => [...userKeys.lists(), filters] as const,
}
