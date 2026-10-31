import { useAuth } from '@/entities/user'
import { LimitSettings } from '@/features/manage-contact-limits'
import { PageHeader } from '@/shared/ui/page-header'

export function ContactLimitsPage() {
  const { role } = useAuth()
  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title="Лимиты контактов"
        description="Сколько раз гостя можно побеспокоить за период"
      />
      <LimitSettings canEdit={role === 'marketer'} />
    </div>
  )
}
