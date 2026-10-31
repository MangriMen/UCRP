import { useState } from 'react'
import { Link, useSearchParams } from 'react-router'
import {
  subjectFromParams,
  subjectParams,
  useAccessPermissions,
  useEffectiveAccess,
  type AccessPermission,
} from '@/entities/access'
import { useAuth } from '@/entities/user'
import {
  SubjectPicker,
  fromSubject,
  toSubject,
  useSubjectDirectory,
  type SubjectValue,
} from '@/features/manage-access-grants'
import { Button } from '@/shared/ui/button'
import { DataTable, type Column } from '@/shared/ui/data-table'
import { EmptyState } from '@/shared/ui/empty-state'
import { PageHeader } from '@/shared/ui/page-header'

interface PermissionRow {
  code: string
  permission?: AccessPermission
}

const kindLabels: Record<AccessPermission['kind'], string> = {
  action: 'Действие',
  metric: 'Показатель',
  field: 'Поле ответа',
}

const columns: Column<PermissionRow>[] = [
  {
    id: 'title',
    header: 'Право',
    cell: (row) => (
      <div className="flex flex-col">
        <span>{row.permission?.title ?? row.code}</span>
        <code className="text-xs text-muted-foreground">{row.code}</code>
      </div>
    ),
  },
  {
    id: 'kind',
    header: 'Вид',
    cell: (row) => (row.permission ? kindLabels[row.permission.kind] : '—'),
  },
  {
    id: 'minK',
    header: 'Порог k',
    cell: (row) => row.permission?.minK ?? '—',
    className: 'text-right',
  },
]

export function AccessEffectivePage() {
  const { can } = useAuth()
  const [searchParams, setSearchParams] = useSearchParams()
  const directory = useSubjectDirectory()
  const permissions = useAccessPermissions()
  const subject = subjectFromParams(searchParams)
  const [draft, setDraft] = useState<SubjectValue | null>(null)
  const effective = useEffectiveAccess(subject)

  const value =
    draft ?? (subject ? fromSubject(subject, directory.users) : { kind: 'tenant', id: '' })

  const choose = (next: SubjectValue) => {
    setDraft(next)
    const target = toSubject(next)
    setSearchParams(target ? subjectParams(target) : {}, { replace: true })
  }

  const rows: PermissionRow[] = (effective.data?.permissions ?? []).map((code) => ({
    code,
    permission: permissions.data?.find((permission) => permission.code === code),
  }))
  const subjectQuery = subject && subjectParams(subject)

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title="Эффективный доступ"
        description="Какие права субъект получает сейчас с учётом потолка роли и всех действующих грантов"
        breadcrumbs={[{ label: 'Доступы', to: '/access/grants' }, { label: 'Эффективный доступ' }]}
      />
      <SubjectPicker value={value} onChange={choose} directory={directory} />
      {!subjectQuery && (
        <EmptyState title="Выберите субъекта" description="Арендатора, пользователя или роль" />
      )}
      {subjectQuery && (
        <>
          <div className="flex flex-wrap items-center justify-between gap-3">
            <p className="text-sm text-muted-foreground">
              {effective.data
                ? `${directory.subjectLabel(subject)}: прав ${effective.data.permissions.length}, версия политики ${effective.data.policyVersion}.${effective.data.userGrantsIncluded ? ' Учтены личные гранты пользователя.' : ''}`
                : directory.subjectLabel(subject)}
            </p>
            <div className="flex flex-wrap gap-2">
              <Button
                variant="outline"
                nativeButton={false}
                render={<Link to={`/access/grants?${subjectQuery}`} />}
              >
                Гранты субъекта
              </Button>
              {can('access.grant.write') && (
                <Button
                  nativeButton={false}
                  render={<Link to={`/access/grants/new?${subjectQuery}`} />}
                >
                  Выдать доступ
                </Button>
              )}
            </div>
          </div>
          <DataTable
            columns={columns}
            rows={effective.data ? rows : undefined}
            getRowId={(row) => row.code}
            isLoading={effective.isLoading}
            error={effective.error}
            emptyTitle="У субъекта нет прав"
          />
        </>
      )}
    </div>
  )
}
