import { useQuery } from '@tanstack/react-query'
import { api } from '@/shared/api/api'
import { ApiError, ensureAccessToken } from '@/shared/api/client'
import { userKeys } from '../model/keys'
import type { Me } from '../model/types'

async function fetchMe(): Promise<Me | null> {
  if (!(await ensureAccessToken())) return null
  try {
    return await api.get('/auth/me')
  } catch (e) {
    if (e instanceof ApiError && e.status === 401) return null
    throw e
  }
}

export function useMe() {
  return useQuery({
    queryKey: userKeys.me,
    queryFn: fetchMe,
    staleTime: Infinity,
    retry: false,
  })
}
