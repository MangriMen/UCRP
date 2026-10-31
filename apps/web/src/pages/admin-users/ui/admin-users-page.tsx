import { useAuth } from '@/entities/user'
import { InviteUserButton } from '@/features/invite-user'
import { UsersManager } from '@/features/manage-users'
import { PageHeader } from '@/shared/ui/page-header'

export function AdminUsersPage() {
  const { can } = useAuth()
  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title="Пользователи"
        description="Учётные записи, роли, приглашения и восстановление доступа"
        actions={can('user.write') && <InviteUserButton />}
      />
      <UsersManager />
    </div>
  )
}
