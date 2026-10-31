import type { components, paths } from './schema'

export type { components, paths }
export type Schema<K extends keyof components['schemas']> = components['schemas'][K]
