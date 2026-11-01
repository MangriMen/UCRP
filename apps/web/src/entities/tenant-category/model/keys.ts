export const tenantCategoryKeys = {
  all: ['tenant-categories'] as const,
  list: (includeInactive: boolean) => ['tenant-categories', { includeInactive }] as const,
  relations: (categoryId: string) => ['tenant-categories', categoryId, 'relations'] as const,
}
