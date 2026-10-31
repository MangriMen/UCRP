import { useMutation, useQueryClient } from '@tanstack/react-query'
import {
  contactLimitKeys,
  type ContactLimitCreateRequest,
  type ContactLimitUpdateRequest,
} from '@/entities/contact-limit'
import { api } from '@/shared/api/api'

function useInvalidateLimits() {
  const queryClient = useQueryClient()
  return () => queryClient.invalidateQueries({ queryKey: contactLimitKeys.all })
}

export function useCreateContactLimit() {
  const invalidate = useInvalidateLimits()
  return useMutation({
    mutationFn: (body: ContactLimitCreateRequest) => api.post('/contact-limits', { body }),
    onSuccess: invalidate,
  })
}

export function useUpdateContactLimit() {
  const invalidate = useInvalidateLimits()
  return useMutation({
    mutationFn: ({ id, body }: { id: string; body: ContactLimitUpdateRequest }) =>
      api.patch('/contact-limits/{id}', { params: { id }, body }),
    onSuccess: invalidate,
  })
}

export function useDeleteContactLimit() {
  const invalidate = useInvalidateLimits()
  return useMutation({
    mutationFn: (id: string) => api.delete('/contact-limits/{id}', { params: { id } }),
    onSuccess: invalidate,
  })
}
