import { useMutation, useQueryClient } from '@tanstack/react-query'
import {
  accessKeys,
  type AccessGrantPatch,
  type AccessGrantRequest,
  type GrantTransition,
} from '@/entities/access'
import { api } from '@/shared/api/api'

function useInvalidateAccess() {
  const queryClient = useQueryClient()
  return () => queryClient.invalidateQueries({ queryKey: accessKeys.all })
}

export function useCreateGrant() {
  const invalidate = useInvalidateAccess()
  return useMutation({
    mutationFn: (body: AccessGrantRequest) => api.post('/access/grants', { body }),
    onSuccess: invalidate,
  })
}

export function useUpdateGrant() {
  const invalidate = useInvalidateAccess()
  return useMutation({
    mutationFn: ({ id, body }: { id: string; body: AccessGrantPatch }) =>
      api.patch('/access/grants/{id}', { params: { id }, body }),
    onSuccess: invalidate,
  })
}

const transitionPaths = {
  activate: '/access/grants/{id}/activate',
  suspend: '/access/grants/{id}/suspend',
  revoke: '/access/grants/{id}/revoke',
} as const

export function useGrantTransition() {
  const invalidate = useInvalidateAccess()
  return useMutation({
    mutationFn: ({
      id,
      transition,
      reason,
    }: {
      id: string
      transition: GrantTransition
      reason: string
    }) => api.post(transitionPaths[transition], { params: { id }, body: { reason } }),
    onSuccess: invalidate,
  })
}

export function useSimulateAccess() {
  return useMutation({
    mutationFn: (body: AccessGrantRequest) => api.post('/access/simulate', { body }),
  })
}
