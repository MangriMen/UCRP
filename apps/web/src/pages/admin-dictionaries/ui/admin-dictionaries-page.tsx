import { useAuth } from '@/entities/user'
import { TenantCategoriesManager } from '@/features/manage-tenant-categories'
import { PageHeader } from '@/shared/ui/page-header'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/shared/ui/tabs'

export function AdminDictionariesPage() {
  const { role } = useAuth()
  const canEdit = role === 'analyst'

  return (
    <div className="flex flex-col gap-6">
      <PageHeader title="Справочники" description="Категории, арендаторы, проекты и типы" />
      <Tabs defaultValue="categories">
        <TabsList>
          <TabsTrigger value="categories">Категории</TabsTrigger>
        </TabsList>
        <TabsContent value="categories">
          <TenantCategoriesManager canEdit={canEdit} />
        </TabsContent>
      </Tabs>
    </div>
  )
}
