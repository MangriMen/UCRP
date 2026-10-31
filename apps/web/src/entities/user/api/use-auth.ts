import { useMe } from './use-me'

export function useAuth() {
  const { data, isPending } = useMe()
  return {
    user: data ?? null,
    role: data?.role ?? null,
    isLoading: isPending,
  }
}
