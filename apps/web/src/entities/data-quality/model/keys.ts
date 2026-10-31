export const dataQualityKeys = {
  all: ['data-quality'] as const,
  list: () => [...dataQualityKeys.all, 'sources'] as const,
  detail: (id: string) => [...dataQualityKeys.all, 'source', id] as const,
}
