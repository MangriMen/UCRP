import { useAuth } from '@/entities/user'
import { DataSourceTypesManager } from '@/features/manage-data-source-types'
import { EventTypesManager } from '@/features/manage-event-types'
import { MallsManager } from '@/features/manage-malls'
import { ProjectsManager } from '@/features/manage-projects'
import { TenantCategoriesManager } from '@/features/manage-tenant-categories'
import { TenantsManager } from '@/features/manage-tenants'
import { PageHeader } from '@/shared/ui/page-header'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/shared/ui/tabs'

export function AdminDictionariesPage() {
  const { role } = useAuth()
  const canEdit = role === 'analyst'

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title="Справочники"
        description="Категории, арендаторы, торговые центры, проекты и типы событий и источников"
      />
      <Tabs defaultValue="categories">
        <TabsList>
          <TabsTrigger value="categories">Категории</TabsTrigger>
          <TabsTrigger value="tenants">Арендаторы</TabsTrigger>
          <TabsTrigger value="malls">Торговые центры</TabsTrigger>
          <TabsTrigger value="projects">Проекты</TabsTrigger>
          <TabsTrigger value="event-types">Типы событий</TabsTrigger>
          <TabsTrigger value="data-source-types">Типы источников</TabsTrigger>
        </TabsList>
        <TabsContent value="categories">
          <TenantCategoriesManager canEdit={canEdit} />
        </TabsContent>
        <TabsContent value="tenants">
          <TenantsManager canEdit={canEdit} />
        </TabsContent>
        <TabsContent value="malls">
          <MallsManager canEdit={canEdit} />
        </TabsContent>
        <TabsContent value="projects">
          <ProjectsManager canEdit={canEdit} />
        </TabsContent>
        <TabsContent value="event-types">
          <EventTypesManager canEdit={canEdit} />
        </TabsContent>
        <TabsContent value="data-source-types">
          <DataSourceTypesManager canEdit={canEdit} />
        </TabsContent>
      </Tabs>
    </div>
  )
}
