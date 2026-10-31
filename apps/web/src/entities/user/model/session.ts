import type { QueryClient } from '@tanstack/react-query'
import { setAccessToken } from '@/shared/api/client'
import { userKeys } from './keys'
import type { Me } from './types'

export function startSession(queryClient: QueryClient, accessToken: string, user: Me) {
  setAccessToken(accessToken)
  queryClient.setQueryData(userKeys.me, user)
}

export function endSession(queryClient: QueryClient) {
  setAccessToken(null)
  queryClient.removeQueries({ predicate: (query) => query.queryKey[0] !== userKeys.me[0] })
  queryClient.setQueryData(userKeys.me, null)
}
