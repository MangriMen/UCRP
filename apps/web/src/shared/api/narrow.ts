export function narrow<T, U extends T>(value: T, guard: (candidate: T) => candidate is U): U {
  if (!guard(value)) throw new Error('Unexpected response shape')
  return value
}

export const hasKey =
  <K extends string>(key: K) =>
  <T extends object>(value: T): value is Extract<T, Record<K, unknown>> =>
    key in value
