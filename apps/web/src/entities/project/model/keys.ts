export interface ProjectListFilters {
  tenantId?: string
  page?: number
  pageSize?: number
}

export const projectKeys = {
  all: ['projects'] as const,
  lists: () => [...projectKeys.all, 'list'] as const,
  list: (filters: ProjectListFilters) => [...projectKeys.lists(), filters] as const,
}
