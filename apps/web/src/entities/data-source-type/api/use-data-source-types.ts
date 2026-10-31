import { useQuery } from '@tanstack/react-query'
import { api } from '@/shared/api/api'
import { dataSourceTypeKeys } from '../model/keys'

export function useDataSourceTypes() {
  return useQuery({
    queryKey: dataSourceTypeKeys.all,
    queryFn: () => api.get('/data-source-types'),
  })
}
