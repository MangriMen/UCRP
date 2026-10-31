import { useMutation, useQueryClient } from '@tanstack/react-query'
import { projectKeys, type ProjectRequest } from '@/entities/project'
import { api } from '@/shared/api/api'

function useInvalidateProjects() {
  const queryClient = useQueryClient()
  return () => queryClient.invalidateQueries({ queryKey: projectKeys.all })
}

export function useCreateProject() {
  const invalidate = useInvalidateProjects()
  return useMutation({
    mutationFn: (body: ProjectRequest) => api.post('/projects', { body }),
    onSuccess: invalidate,
  })
}

export function useUpdateProject() {
  const invalidate = useInvalidateProjects()
  return useMutation({
    mutationFn: ({ id, body }: { id: string; body: ProjectRequest }) =>
      api.patch('/projects/{id}', { params: { id }, body }),
    onSuccess: invalidate,
  })
}

export function useDeleteProject() {
  const invalidate = useInvalidateProjects()
  return useMutation({
    mutationFn: (id: string) => api.delete('/projects/{id}', { params: { id } }),
    onSuccess: invalidate,
  })
}
