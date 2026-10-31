import { UsersManager } from '@/features/manage-users'
import { PageHeader } from '@/shared/ui/page-header'

export function AdminUsersPage() {
  return (
    <div className="flex flex-col gap-6">
      <PageHeader title="Пользователи" description="Учётные записи, роли и привязка к арендатору" />
      <UsersManager />
    </div>
  )
}
