import { useCallback, useEffect, useRef, useState } from 'react'
import type { ListParams } from '../list-params'

const SEARCH_DEBOUNCE_MS = 300

export function useDebouncedSearchParam(
  { get, update }: Pick<ListParams, 'get' | 'update'>,
  key: string,
) {
  const [value, setValue] = useState(get(key))
  const timer = useRef<ReturnType<typeof setTimeout>>(undefined)

  useEffect(() => () => clearTimeout(timer.current), [])

  const onChange = (next: string) => {
    setValue(next)
    clearTimeout(timer.current)
    timer.current = setTimeout(() => update({ [key]: next.trim() }), SEARCH_DEBOUNCE_MS)
  }

  const clear = useCallback(() => {
    clearTimeout(timer.current)
    setValue('')
  }, [])

  return { value, onChange, clear }
}
