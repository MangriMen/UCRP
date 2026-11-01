import { useEffect } from 'react'
import type { ListParams } from '../list-params'
import { ALL_OPTION, optionParam } from '../search-params'

export function useOptionParam(
  { get, update }: Pick<ListParams, 'get' | 'update'>,
  key: string,
  knownValues: string[] | undefined,
) {
  const current = get(key)
  const isUnknown = current !== '' && knownValues !== undefined && !knownValues.includes(current)

  useEffect(() => {
    if (isUnknown) update({ [key]: null })
  }, [isUnknown, key, update])

  return {
    value: isUnknown ? ALL_OPTION : current || ALL_OPTION,
    onChange: (next: string) => update({ [key]: optionParam(next) }),
  }
}
