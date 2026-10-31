import type { components, paths } from './schema'

export type { paths }
export type Schema<K extends keyof components['schemas']> = components['schemas'][K]
