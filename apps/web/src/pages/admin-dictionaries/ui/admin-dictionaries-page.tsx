import { useSearchParams } from 'react-router'
import { useAuth } from '@/entities/user'
import { DataSourceTypesManager } from '@/features/manage-data-source-types'
import { EventTypesManager } from '@/features/manage-event-types'
import { MallsManager } from '@/features/manage-malls'
import { ProjectsManager } from '@/features/manage-projects'
import { TenantCategoriesManager } from '@/features/manage-tenant-categories'
import { ButtonLink } from '@/shared/ui/button'
import { PageHeader } from '@/shared/ui/page-header'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/shared/ui/tabs'

const RELATIONS_PARAM = 'relations'

export function AdminDictionariesPage() {
  const { can } = useAuth()
  const [searchParams, setSearchParams] = useSearchParams()
  const setRelationsCategory = (categoryId: string | null) =>
    setSearchParams(
      (current) => {
        const params = new URLSearchParams(current)
        if (categoryId) params.set(RELATIONS_PARAM, categoryId)
        else params.delete(RELATIONS_PARAM)
        return params
      },
      { replace: true },
    )
  const canEditReference = can('reference.write')
  const canEditSourceTypes = can('data_source.write')

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
          <TenantCategoriesManager
            canEdit={canEditReference}
            relationsCategoryId={searchParams.get(RELATIONS_PARAM)}
            onRelationsCategoryChange={setRelationsCategory}
          />
        </TabsContent>
        <TabsContent value="tenants" className="flex flex-col items-start gap-3 pt-4">
          <p className="text-sm text-muted-foreground">
            Арендаторы с показателями, источниками и настройками вынесены в отдельный раздел.
          </p>
          <ButtonLink to="/tenants">Открыть арендаторов</ButtonLink>
        </TabsContent>
        <TabsContent value="malls">
          <MallsManager canEdit={canEditReference} />
        </TabsContent>
        <TabsContent value="projects">
          <ProjectsManager canEdit={canEditReference} />
        </TabsContent>
        <TabsContent value="event-types">
          <EventTypesManager canEdit={canEditReference} />
        </TabsContent>
        <TabsContent value="data-source-types">
          <DataSourceTypesManager canEdit={canEditSourceTypes} />
        </TabsContent>
      </Tabs>
    </div>
  )
}
