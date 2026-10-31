export const tenantCategoryKeys = {
  all: ['tenant-categories'] as const,
  relations: (categoryId: string) => ['tenant-categories', categoryId, 'relations'] as const,
}
