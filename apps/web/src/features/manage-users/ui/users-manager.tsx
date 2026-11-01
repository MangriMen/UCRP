import { useState } from 'react'
import { roleLabels, useAuth, useUsers, type Role, type User } from '@/entities/user'
import { formatDateTime } from '@/shared/lib/format'
import { useDebouncedSearchParam } from '@/shared/lib/hooks/use-debounced-search-param'
import { useListParams } from '@/shared/lib/list-params'
import { ALL_OPTION, optionParam } from '@/shared/lib/search-params'
import { ActiveBadge } from '@/shared/ui/active-badge'
import { Alert, AlertDescription, AlertTitle } from '@/shared/ui/alert'
import { Button } from '@/shared/ui/button'
import { ConfirmDialog } from '@/shared/ui/confirm-dialog'
import { DataTable, type Column } from '@/shared/ui/data-table'
import { ErrorBanner } from '@/shared/ui/error-banner'
import { Input } from '@/shared/ui/input'
import { OptionSelect } from '@/shared/ui/option-select'
import { assignableRoles, recoverableRoles } from '../model/schema'
import { useDeleteUser, useRecoverUser } from '../api/use-user-mutations'
import { RenameUserDialog } from './rename-user-dialog'

const WRITE_PERMISSION = 'user.write'

const roleFilterItems = [
  { value: ALL_OPTION, label: 'Все роли' },
  ...assignableRoles.map((role) => ({ value: role, label: roleLabels[role] })),
]

const isRole = (value: string): value is Role => value in roleLabels

const canRecover = (user: User) => user.isActive && recoverableRoles.includes(user.role)

export function UsersManager() {
  const params = useListParams()
  const { get, update, page, pageSize, setPage, setPageSize, reset } = params
  const search = useDebouncedSearchParam(params, 'q')
  const role = get('role')
  const isFiltered = params.isFiltered(['q', 'role'])

  const users = useUsers({
    q: get('q') || undefined,
    role: isRole(role) ? role : undefined,
    page,
    pageSize,
  })
  const { user: me, can } = useAuth()
  const canWrite = can(WRITE_PERMISSION)
  const remove = useDeleteUser()
  const recover = useRecoverUser()
  const [renaming, setRenaming] = useState<User | null>(null)
  const [deleting, setDeleting] = useState<User | null>(null)
  const [recovering, setRecovering] = useState<User | null>(null)
  const [recoveredEmail, setRecoveredEmail] = useState<string | null>(null)

  const resetFilters = () => {
    search.clear()
    reset()
  }

  const confirmRecovery = () => {
    if (!recovering) return
    const { id, email } = recovering
    setRecoveredEmail(null)
    recover.mutate(id, {
      onSuccess: () => setRecoveredEmail(email),
      onSettled: () => setRecovering(null),
    })
  }

  const actionsColumn: Column<User> = {
    id: 'actions',
    header: '',
    className: 'text-right',
    cell: (u) => (
      <div className="flex flex-wrap justify-end gap-2">
        <Button variant="outline" size="sm" onClick={() => setRenaming(u)}>
          Изменить
        </Button>
        {canRecover(u) && (
          <Button variant="outline" size="sm" onClick={() => setRecovering(u)}>
            Восстановить доступ
          </Button>
        )}
        <Button
          variant="outline"
          size="sm"
          disabled={!u.isActive || u.id === me?.userId}
          onClick={() => setDeleting(u)}
        >
          Деактивировать
        </Button>
      </div>
    ),
  }

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
    ...(canWrite ? [actionsColumn] : []),
  ]

  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-wrap items-center gap-2">
        <Input
          value={search.value}
          onChange={(event) => search.onChange(event.target.value)}
          placeholder="Email или имя"
          aria-label="Поиск пользователей"
          className="w-full sm:w-64"
        />
        <OptionSelect
          items={roleFilterItems}
          value={role || ALL_OPTION}
          onValueChange={(next) => update({ role: optionParam(next) })}
          ariaLabel="Роль"
          className="w-full sm:w-48"
        />
      </div>
      {remove.error && <ErrorBanner error={remove.error} />}
      {recover.error && <ErrorBanner error={recover.error} />}
      {recoveredEmail && (
        <Alert>
          <AlertTitle>Ссылка отправлена</AlertTitle>
          <AlertDescription>
            {`Ссылка для восстановления доступа отправлена на ${recoveredEmail}. Она действует 24 часа.`}
          </AlertDescription>
        </Alert>
      )}
      <DataTable
        columns={columns}
        rows={users.data?.items}
        getRowId={(u) => u.id}
        isLoading={users.isLoading}
        error={users.error}
        emptyTitle={isFiltered ? 'Пользователи по фильтрам не найдены' : 'Пользователей пока нет'}
        onResetFilters={isFiltered ? resetFilters : undefined}
        meta={users.data?.meta}
        onPageChange={setPage}
        onPageSizeChange={setPageSize}
      />
      {renaming && (
        <RenameUserDialog user={renaming} onOpenChange={(open) => !open && setRenaming(null)} />
      )}
      <ConfirmDialog
        open={recovering !== null}
        onOpenChange={(open) => !open && setRecovering(null)}
        title="Восстановить доступ?"
        description={`На ${recovering?.email ?? ''} придёт одноразовая ссылка для установки нового пароля. После её использования старый пароль перестанет действовать.`}
        confirmLabel="Отправить ссылку"
        isPending={recover.isPending}
        onConfirm={confirmRecovery}
      />
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
