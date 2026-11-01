import type { QueryClient } from '@tanstack/react-query'
import { setAccessToken } from '@/shared/api/client'
import { userKeys } from './keys'
import type { Me } from './types'

export function startSession(queryClient: QueryClient, accessToken: string, user: Me) {
  setAccessToken(accessToken)
  queryClient.setQueryData(userKeys.me, user)
}

let endedSessionUserId: string | null = null

export function endSession(queryClient: QueryClient) {
  endedSessionUserId =
    queryClient.getQueryData<Me | null>(userKeys.me)?.userId ?? endedSessionUserId
  setAccessToken(null)
  queryClient.removeQueries({ predicate: (query) => query.queryKey[0] !== userKeys.me[0] })
  queryClient.setQueryData(userKeys.me, null)
}

export function returnPathFor(user: Me, from: string) {
  return endedSessionUserId === null || endedSessionUserId === user.userId ? from : '/'
}
