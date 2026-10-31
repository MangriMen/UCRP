import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { api } from '@/shared/api/api'
import { hasKey, narrow } from '@/shared/api/narrow'
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
        query: {
          q: filters.query && filters.query.length >= MIN_SEARCH_LENGTH ? filters.query : undefined,
          source_id: filters.sourceId,
          tenant_id: filters.tenantId,
          period_from: filters.periodFrom,
          period_to: filters.periodTo,
          page: filters.page,
          page_size: filters.pageSize,
        },
      }),
    placeholderData: keepPreviousData,
  })
}

export function useGuest(id: string) {
  return useQuery({
    queryKey: guestKeys.detail(id),
    queryFn: () => api.get('/guests/{id}', { params: { id } }),
    select: (data): GuestProfile => narrow(data, hasKey('tenantCount')),
  })
}

export function useGuestEvents(id: string, filters: GuestEventFilters = {}) {
  return useQuery({
    queryKey: guestKeys.events(id, filters),
    queryFn: () =>
      api.get('/guests/{id}/events', {
        params: { id },
        query: {
          limit: filters.pageSize,
          offset:
            filters.page && filters.pageSize ? (filters.page - 1) * filters.pageSize : undefined,
        },
      }),
    select: (data) => narrow(data, hasKey('meta')),
    placeholderData: keepPreviousData,
  })
}

export function useGuestMovement(id: string) {
  return useQuery({
    queryKey: guestKeys.movement(id),
    queryFn: () => api.get('/guests/{id}/movement', { params: { id } }),
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
