export interface ImportListFilters {
  dataSourceId?: string
  status?: string
  page?: number
  pageSize?: number
}

export const importKeys = {
  all: ['imports'] as const,
  lists: () => [...importKeys.all, 'list'] as const,
  list: (filters: ImportListFilters) => [...importKeys.lists(), filters] as const,
  detail: (id: string) => [...importKeys.all, 'detail', id] as const,
  rejectedRows: (id: string, page: number, pageSize: number) =>
    [...importKeys.detail(id), 'rejected-rows', { page, pageSize }] as const,
}
