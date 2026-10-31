import type { components, paths } from '@ucrp/api-types'

export type { paths }
export type Schema<K extends keyof components['schemas']> = components['schemas'][K]
