import { useQuery } from '@tanstack/react-query'
import { api } from '@/shared/api/api'
import { dataQualityKeys } from '../model/keys'

export function useDataQualitySources() {
  return useQuery({
    queryKey: dataQualityKeys.list(),
    queryFn: () => api.get('/data-quality/sources'),
  })
}

export function useDataQualitySource(id: string | null) {
  return useQuery({
    queryKey: dataQualityKeys.detail(id ?? ''),
    queryFn: () => api.get('/data-quality/sources/{id}', { params: { id: id ?? '' } }),
    enabled: id !== null,
  })
}
