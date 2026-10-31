import { useState } from 'react'
import { roleLabels, useUsers, type Role, type User } from '@/entities/user'
import { formatDateTime } from '@/shared/lib/format'
import { ActiveBadge } from '@/shared/ui/active-badge'
import { Button } from '@/shared/ui/button'
import { ConfirmDialog } from '@/shared/ui/confirm-dialog'
import { DataTable, type Column } from '@/shared/ui/data-table'
import { ErrorBanner } from '@/shared/ui/error-banner'
import { Input } from '@/shared/ui/input'
import { OptionSelect } from '@/shared/ui/option-select'
import { assignableRoles } from '../model/schema'
import { useDeleteUser } from '../api/use-user-mutations'
import { UserFormDialog } from './user-form-dialog'

const ALL = 'all'

const roleFilterItems = [
  { value: ALL, label: 'Все роли' },
  ...assignableRoles.map((role) => ({ value: role, label: roleLabels[role] })),
]

const isRole = (value: string): value is Role => value in roleLabels

interface FormState {
  user: User | null
}

export function UsersManager() {
  const [page, setPage] = useState(1)
  const [search, setSearch] = useState('')
  const [roleFilter, setRoleFilter] = useState(ALL)
  const users = useUsers({
    q: search.trim() || undefined,
    role: isRole(roleFilter) ? roleFilter : undefined,
    page,
  })
  const remove = useDeleteUser()
  const [form, setForm] = useState<FormState | null>(null)
  const [deleting, setDeleting] = useState<User | null>(null)

  const columns: Column<User>[] = [
    { id: 'name', header: 'Имя', cell: (u) => u.displayName },
    { id: 'email', header: 'Email', cell: (u) => u.email },
    { id: 'role', header: 'Роль', cell: (u) => roleLabels[u.role] },
    {
      id: 'lastLogin',
      header: 'Последний вход',
      cell: (u) => (u.lastLoginAt ? formatDateTime(u.lastLoginAt) : '—'),
    },
    { id: 'status', header: 'Статус', cell: (u) => <ActiveBadge isActive={u.isActive} /> },
    {
      id: 'actions',
      header: '',
      className: 'text-right',
      cell: (u) => (
        <div className="flex justify-end gap-2">
          <Button variant="outline" size="sm" onClick={() => setForm({ user: u })}>
            Изменить
          </Button>
          <Button variant="outline" size="sm" disabled={!u.isActive} onClick={() => setDeleting(u)}>
            Деактивировать
          </Button>
        </div>
      ),
    },
  ]

  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <div className="flex flex-wrap gap-2">
          <Input
            value={search}
            onChange={(event) => {
              setSearch(event.target.value)
              setPage(1)
            }}
            placeholder="Email или имя"
            aria-label="Поиск пользователей"
            className="w-64"
          />
          <OptionSelect
            items={roleFilterItems}
            value={roleFilter}
            onValueChange={(next) => {
              setRoleFilter(next)
              setPage(1)
            }}
            ariaLabel="Роль"
            className="w-48"
          />
        </div>
        <Button onClick={() => setForm({ user: null })}>Добавить пользователя</Button>
      </div>
      {remove.error && <ErrorBanner error={remove.error} />}
      <DataTable
        columns={columns}
        rows={users.data?.items}
        getRowId={(u) => u.id}
        isLoading={users.isLoading}
        error={users.error}
        emptyTitle="Пользователи не найдены"
        meta={users.data?.meta}
        onPageChange={setPage}
      />
      {form && (
        <UserFormDialog open user={form.user} onOpenChange={(open) => !open && setForm(null)} />
      )}
      <ConfirmDialog
        open={deleting !== null}
        onOpenChange={(open) => !open && setDeleting(null)}
        title="Деактивировать пользователя?"
        description={`«${deleting?.displayName ?? ''}» потеряет доступ к системе.`}
        confirmLabel="Деактивировать"
        destructive
        isPending={remove.isPending}
        onConfirm={() => {
          if (deleting) remove.mutate(deleting.id, { onSettled: () => setDeleting(null) })
        }}
      />
    </div>
  )
}
