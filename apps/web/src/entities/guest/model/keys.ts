export interface GuestSearchFilters {
  query?: string
  sourceId?: string
  tenantId?: string
  periodFrom?: string
  periodTo?: string
  page?: number
  pageSize?: number
}

export interface GuestEventFilters {
  from?: string
  to?: string
  tenantId?: string
  page?: number
  pageSize?: number
}

export interface GuestContactFilters {
  page?: number
  pageSize?: number
}

export const MIN_SEARCH_LENGTH = 2

export const guestKeys = {
  all: ['guests'] as const,
  lists: () => [...guestKeys.all, 'list'] as const,
  list: (filters: GuestSearchFilters) => [...guestKeys.lists(), filters] as const,
  detail: (id: string) => [...guestKeys.all, 'detail', id] as const,
  events: (id: string, filters: GuestEventFilters) =>
    [...guestKeys.detail(id), 'events', filters] as const,
  movement: (id: string) => [...guestKeys.detail(id), 'movement'] as const,
  contacts: (id: string, filters: GuestContactFilters) =>
    [...guestKeys.detail(id), 'contacts', filters] as const,
}
