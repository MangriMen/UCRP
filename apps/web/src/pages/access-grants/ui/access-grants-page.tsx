import { Link, useNavigate } from 'react-router'
import {
  GRANT_SUBJECT_TYPES,
  GrantStatusBadge,
  grantStatusLabels,
  parseSubjectType,
  presetLabel,
  subjectTypeLabels,
  useAccessGrants,
  type AccessGrant,
  type GrantStatus,
  type GrantSubjectType,
} from '@/entities/access'
import { useAuth } from '@/entities/user'
import { useSubjectDirectory, type SubjectDirectory } from '@/features/manage-access-grants'
import { formatDateTime } from '@/shared/lib/format'
import { useListParams } from '@/shared/lib/list-params'
import { ALL_OPTION, optionParam } from '@/shared/lib/search-params'
import { Button } from '@/shared/ui/button'
import { DataTable, type Column } from '@/shared/ui/data-table'
import { OptionSelect, type Option } from '@/shared/ui/option-select'
import { PageHeader } from '@/shared/ui/page-header'

const FILTER_KEYS = ['status', 'subject_type', 'subject_id'] as const

const STATUSES = Object.keys(grantStatusLabels) as GrantStatus[]

const statusItems: Option[] = [
  { value: ALL_OPTION, label: 'Любой статус' },
  ...STATUSES.map((status) => ({ value: status, label: grantStatusLabels[status] })),
]

const subjectTypeItems: Option[] = [
  { value: ALL_OPTION, label: 'Любой субъект' },
  ...GRANT_SUBJECT_TYPES.map((type) => ({ value: type, label: subjectTypeLabels[type] })),
]

const pick = <T extends string>(values: readonly T[], value: string) =>
  values.find((candidate) => candidate === value)

function subjectItems(type: GrantSubjectType | undefined, directory: SubjectDirectory): Option[] {
  const all = { value: ALL_OPTION, label: 'Все' }
  if (type === 'tenant')
    return [all, ...directory.tenants.map(({ id, label }) => ({ value: id, label }))]
  if (type === 'user')
    return [all, ...directory.users.map(({ id, label }) => ({ value: id, label }))]
  if (type === 'role') return [all, { value: 'tenant', label: 'Все арендаторы' }]
  return []
}

const period = (grant: AccessGrant) =>
  `${formatDateTime(grant.validFrom)} — ${grant.validTo ? formatDateTime(grant.validTo) : 'бессрочно'}`

export function AccessGrantsPage() {
  const navigate = useNavigate()
  const { can } = useAuth()
  const directory = useSubjectDirectory()
  const { get, update, page, pageSize, setPage, setPageSize, reset } = useListParams()
  const status = pick(STATUSES, get('status'))
  const subjectType = parseSubjectType(get('subject_type'))
  const subjectId = get('subject_id')
  const isFiltered = FILTER_KEYS.some((key) => get(key) !== '')

  const grants = useAccessGrants({
    status,
    subjectType,
    subjectId: subjectType ? subjectId || undefined : undefined,
    page,
    pageSize,
  })

  const columns: Column<AccessGrant>[] = [
    {
      id: 'subject',
      header: 'Субъект',
      cell: (grant) => (
        <div className="flex flex-col">
          <span className="font-medium">{directory.subjectLabel(grant)}</span>
          <span className="text-xs text-muted-foreground">
            {subjectTypeLabels[grant.subjectType]}
          </span>
        </div>
      ),
    },
    {
      id: 'content',
      header: 'Состав',
      cell: (grant) =>
        grant.presetCode ? presetLabel(grant.presetCode) : `Прав: ${grant.items.length}`,
    },
    { id: 'period', header: 'Срок', cell: period },
    { id: 'status', header: 'Статус', cell: (grant) => <GrantStatusBadge status={grant.status} /> },
    { id: 'updatedAt', header: 'Изменён', cell: (grant) => formatDateTime(grant.updatedAt) },
  ]

  const select = (key: string) => (value: string) => update({ [key]: optionParam(value) })

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title="Доступы"
        description="Гранты арендаторам и управляющим: кто, что и до какого срока видит"
        actions={
          <div className="flex flex-wrap gap-2">
            <Button variant="outline" nativeButton={false} render={<Link to="/access/effective" />}>
              Эффективный доступ
            </Button>
            {can('audit.read') && (
              <Button variant="outline" nativeButton={false} render={<Link to="/access/audit" />}>
                Аудит доступа
              </Button>
            )}
            {can('access.grant.write') && (
              <Button nativeButton={false} render={<Link to="/access/grants/new" />}>
                Выдать доступ
              </Button>
            )}
          </div>
        }
      />
      <div className="flex flex-wrap items-center gap-3">
        <OptionSelect
          ariaLabel="Статус"
          className="w-48"
          items={statusItems}
          value={status ?? ALL_OPTION}
          onValueChange={select('status')}
        />
        <OptionSelect
          ariaLabel="Тип субъекта"
          className="w-48"
          items={subjectTypeItems}
          value={subjectType ?? ALL_OPTION}
          onValueChange={(value) => update({ subject_type: optionParam(value), subject_id: null })}
        />
        {subjectType && (
          <OptionSelect
            ariaLabel="Субъект"
            className="w-64"
            items={subjectItems(subjectType, directory)}
            value={subjectId || ALL_OPTION}
            onValueChange={select('subject_id')}
          />
        )}
        <Button type="button" variant="outline" disabled={!isFiltered} onClick={reset}>
          Сбросить
        </Button>
      </div>
      <DataTable
        columns={columns}
        rows={grants.data?.items}
        getRowId={(grant) => grant.id}
        isLoading={grants.isLoading}
        error={grants.error}
        emptyTitle={isFiltered ? 'Гранты по фильтрам не найдены' : 'Грантов пока нет'}
        emptyAction={
          isFiltered ? (
            <Button type="button" variant="outline" onClick={reset}>
              Сбросить фильтры
            </Button>
          ) : undefined
        }
        meta={grants.data?.meta}
        onPageChange={setPage}
        onPageSizeChange={setPageSize}
        onRowClick={(grant) => navigate(`/access/grants/${grant.id}`)}
      />
    </div>
  )
}
