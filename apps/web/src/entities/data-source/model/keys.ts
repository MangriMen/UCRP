export const dataSourceKeys = {
  all: ['data-sources'] as const,
  list: () => [...dataSourceKeys.all, 'list'] as const,
}
