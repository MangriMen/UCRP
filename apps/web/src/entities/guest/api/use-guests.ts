import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { api } from '@/shared/api/api'
import {
  guestKeys,
  MIN_SEARCH_LENGTH,
  type GuestContactFilters,
  type GuestEventFilters,
  type GuestSearchFilters,
} from '../model/keys'
import type { GuestProfile } from '../model/types'

export function useGuestSearch(filters: GuestSearchFilters) {
  return useQuery({
    queryKey: guestKeys.list(filters),
    queryFn: () =>
      api.get('/guests', {
        query: { q: filters.query, page: filters.page, page_size: filters.pageSize },
      }),
    enabled: filters.query.length >= MIN_SEARCH_LENGTH,
    placeholderData: keepPreviousData,
  })
}

export function useGuest(id: string) {
  return useQuery<GuestProfile>({
    queryKey: guestKeys.detail(id),
    queryFn: () => api.get('/guests/{id}', { params: { id } }),
  })
}

export function useGuestEvents(id: string, filters: GuestEventFilters = {}) {
  return useQuery({
    queryKey: guestKeys.events(id, filters),
    queryFn: () =>
      api.get('/guests/{id}/events', {
        params: { id },
        query: {
          from: filters.from,
          to: filters.to,
          tenant_id: filters.tenantId,
          page: filters.page,
          page_size: filters.pageSize,
        },
      }),
    placeholderData: keepPreviousData,
  })
}

export function useGuestContacts(id: string, filters: GuestContactFilters = {}) {
  return useQuery({
    queryKey: guestKeys.contacts(id, filters),
    queryFn: () =>
      api.get('/guests/{id}/contacts', {
        params: { id },
        query: { page: filters.page, page_size: filters.pageSize },
      }),
    placeholderData: keepPreviousData,
  })
}
