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
import type { GuestEventSort, GuestListSort, GuestProfile } from '../model/types'

export function useGuestSearch(filters: GuestSearchFilters, options: { enabled?: boolean } = {}) {
  return useQuery({
    enabled: options.enabled,
    queryKey: guestKeys.list(filters),
    queryFn: () =>
      api.get('/guests', {
        query: {
          q: filters.query && filters.query.length >= MIN_SEARCH_LENGTH ? filters.query : undefined,
          source_id: filters.sourceId,
          tenant_id: filters.tenantId,
          period_from: filters.periodFrom,
          period_to: filters.periodTo,
          status: filters.status,
          segment_id: filters.segmentId,
          has_purchases: filters.hasPurchases,
          first_seen_from: filters.firstSeenFrom,
          last_seen_to: filters.lastSeenTo,
          sort: filters.sort as GuestListSort | undefined,
          page: filters.page,
          page_size: filters.pageSize,
        },
      }),
    placeholderData: keepPreviousData,
  })
}

export function useTenantGuests(tenantId: string, filters: GuestSearchFilters) {
  return useQuery({
    queryKey: guestKeys.tenantList(tenantId, filters),
    queryFn: () =>
      api.get('/tenants/{id}/guests', {
        params: { id: tenantId },
        query: {
          q: filters.query && filters.query.length >= MIN_SEARCH_LENGTH ? filters.query : undefined,
          source_id: filters.sourceId,
          period_from: filters.periodFrom,
          period_to: filters.periodTo,
          status: filters.status,
          has_purchases: filters.hasPurchases,
          sort: filters.sort as GuestListSort | undefined,
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
          type: filters.type,
          tenant_id: filters.tenantId,
          from: filters.from,
          to: filters.to,
          amount_min: filters.amountMin,
          amount_max: filters.amountMax,
          sort: filters.sort as GuestEventSort | undefined,
          page: filters.page,
          page_size: filters.pageSize,
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
