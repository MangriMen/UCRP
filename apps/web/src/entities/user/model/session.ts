import type { QueryClient } from '@tanstack/react-query'
import { setAccessToken } from '@/shared/api/client'
import { userKeys } from './keys'

export function endSession(queryClient: QueryClient) {
  setAccessToken(null)
  queryClient.removeQueries({ predicate: (query) => query.queryKey[0] !== userKeys.me[0] })
  queryClient.setQueryData(userKeys.me, null)
}
