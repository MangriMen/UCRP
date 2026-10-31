import { useCallback } from 'react'
import { useSearchParams } from 'react-router'

const TAB_PARAM = 'tab'

export function useUrlTab<T extends string>(tabs: readonly T[], fallback: T) {
  const [searchParams, setSearchParams] = useSearchParams()
  const requested = searchParams.get(TAB_PARAM)
  const tab = tabs.find((candidate) => candidate === requested) ?? fallback

  const setTab = useCallback(
    (next: string) =>
      setSearchParams(
        (current) => {
          const params = new URLSearchParams(current)
          if (next === fallback) params.delete(TAB_PARAM)
          else params.set(TAB_PARAM, next)
          return params
        },
        { replace: true },
      ),
    [fallback, setSearchParams],
  )

  return [tab, setTab] as const
}
