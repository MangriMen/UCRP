import type { ReactNode } from 'react'
import { Link, useParams } from 'react-router'
import {
  AccessAuditTable,
  GrantStatusBadge,
  effectLabels,
  presetLabel,
  scopeTypeLabels,
  subjectParams,
  subjectTypeLabels,
  useAccessGrant,
  useAccessPermissions,
  type AccessGrant,
  type AccessGrantItem,
} from '@/entities/access'
import { AuditActor, GrantActions, useSubjectDirectory } from '@/features/manage-access-grants'
import { formatDateTime } from '@/shared/lib/format'
import { useUrlTab } from '@/shared/lib/hooks/use-url-tab'
import { useListParams } from '@/shared/lib/list-params'
import { Button } from '@/shared/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/shared/ui/card'
import { DataTable, type Column } from '@/shared/ui/data-table'
import { EmptyState } from '@/shared/ui/empty-state'
import { ErrorBanner } from '@/shared/ui/error-banner'
import { PageHeader } from '@/shared/ui/page-header'
import { Skeleton } from '@/shared/ui/skeleton'
import { StatusBadge } from '@/shared/ui/status-badge'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/shared/ui/tabs'

const TABS = ['overview', 'history'] as const

function Field({ label, children }: { label: string; children: ReactNode }) {
  return (
    <div className="flex flex-col gap-0.5">
      <dt className="text-xs text-muted-foreground">{label}</dt>
      <dd className="text-sm break-words">{children}</dd>
    </div>
  )
}

function GrantItems({ grant }: { grant: AccessGrant }) {
  const permissions = useAccessPermissions()
  const titleOf = (code: string) => permissions.data?.find((p) => p.code === code)?.title ?? code
  const columns: Column<AccessGrantItem>[] = [
    {
      id: 'permission',
      header: 'Право',
      cell: (item) => (
        <div className="flex flex-col">
          <span>{titleOf(item.permission)}</span>
          <code className="text-xs text-muted-foreground">{item.permission}</code>
        </div>
      ),
    },
    {
      id: 'effect',
      header: 'Эффект',
      cell: (item) => (
        <StatusBadge tone={item.effect === 'allow' ? 'success' : 'danger'}>
          {effectLabels[item.effect]}
        </StatusBadge>
      ),
    },
  ]
  return (
    <DataTable
      columns={columns}
      rows={grant.items}
      getRowId={(item) => item.permission}
      emptyTitle="В гранте нет прав"
    />
  )
}

function GrantHistory({ grantId }: { grantId: string }) {
  const directory = useSubjectDirectory()
  const { page, pageSize, setPage, setPageSize } = useListParams()
  return (
    <AccessAuditTable
      filters={{ grantId, page, pageSize }}
      actorLabel={(entry) => <AuditActor entry={entry} directory={directory} />}
      onPageChange={setPage}
      onPageSizeChange={setPageSize}
      emptyTitle="Изменений гранта пока нет"
    />
  )
}

export function AccessGrantDetailsPage() {
  const { id = '' } = useParams()
  const [tab, setTab] = useUrlTab(TABS, 'overview')
  const directory = useSubjectDirectory()
  const { data: grant, error, isLoading } = useAccessGrant(id)

  if (isLoading) return <Skeleton className="h-40 w-full" />
  if (error) return <ErrorBanner error={error} />
  if (!grant) return <EmptyState title="Грант не найден" />

  const subjectName = directory.subjectLabel(grant)
  const subjectQuery = subjectParams(grant)
  const scopes = grant.scopes

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title={`Грант: ${subjectName}`}
        description={`${subjectTypeLabels[grant.subjectType]} · ${grant.subjectId}`}
        breadcrumbs={[{ label: 'Доступы', to: '/access/grants' }, { label: subjectName }]}
        actions={<GrantActions grant={grant} />}
      />
      <Tabs value={tab} onValueChange={setTab}>
        <TabsList>
          <TabsTrigger value="overview">Обзор</TabsTrigger>
          <TabsTrigger value="history">История</TabsTrigger>
        </TabsList>
        <TabsContent value="overview" className="flex flex-col gap-6 pt-4">
          <Card>
            <CardHeader>
              <CardTitle>Параметры</CardTitle>
            </CardHeader>
            <CardContent>
              <dl className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
                <Field label="Статус">
                  <GrantStatusBadge status={grant.status} />
                </Field>
                <Field label="Срок">
                  {`${formatDateTime(grant.validFrom)} — ${grant.validTo ? formatDateTime(grant.validTo) : 'бессрочно'}`}
                </Field>
                <Field label="Пресет">
                  {grant.presetCode ? presetLabel(grant.presetCode) : '—'}
                </Field>
                <Field label="Причина">{grant.reason || '—'}</Field>
                <Field label="Выдал">{directory.userLabel(grant.createdBy)}</Field>
                <Field label="Версия политики">{grant.policyVersion}</Field>
                <Field label="Создан">{formatDateTime(grant.createdAt)}</Field>
                <Field label="Изменён">{formatDateTime(grant.updatedAt)}</Field>
              </dl>
            </CardContent>
          </Card>
          <section className="flex flex-col gap-3">
            <h2 className="text-base font-semibold">Права</h2>
            <GrantItems grant={grant} />
          </section>
          <section className="flex flex-col gap-3">
            <h2 className="text-base font-semibold">Ограничения</h2>
            {scopes.length === 0 ? (
              <p className="text-sm text-muted-foreground">Без ограничений</p>
            ) : (
              <ul className="flex flex-col gap-1 text-sm">
                {scopes.map((scope) => (
                  <li key={`${scope.type}:${scope.value}`}>
                    {`${scopeTypeLabels[scope.type]}: ${scope.type === 'tenant' ? directory.subjectLabel({ subjectType: 'tenant', subjectId: scope.value }) : scope.value}`}
                  </li>
                ))}
              </ul>
            )}
          </section>
          <div>
            <Button
              variant="outline"
              nativeButton={false}
              render={<Link to={`/access/effective?${subjectQuery}`} />}
            >
              Эффективный доступ субъекта
            </Button>
          </div>
        </TabsContent>
        <TabsContent value="history" className="pt-4">
          <GrantHistory grantId={grant.id} />
        </TabsContent>
      </Tabs>
    </div>
  )
}
