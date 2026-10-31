export interface DataSourceListFilters {
  tenantId?: string
  typeId?: string
  isActive?: boolean
}

export const dataSourceKeys = {
  all: ['data-sources'] as const,
  list: (filters: DataSourceListFilters = {}) => [...dataSourceKeys.all, 'list', filters] as const,
}
