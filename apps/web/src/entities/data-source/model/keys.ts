export interface DataSourceListFilters {
  tenantId?: string
  typeId?: string
  isActive?: boolean
  q?: string
  importStatus?: string
  sort?: string
  page?: number
  pageSize?: number
}

export const dataSourceKeys = {
  all: ['data-sources'] as const,
  list: (filters: DataSourceListFilters = {}) => [...dataSourceKeys.all, 'list', filters] as const,
}
