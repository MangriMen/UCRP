import { useNavigate, useParams, useSearchParams } from 'react-router'
import {
  isGrantEditable,
  subjectFromParams,
  useAccessGrant,
  type AccessGrant,
} from '@/entities/access'
import { GrantForm, useSubjectDirectory } from '@/features/manage-access-grants'
import { EmptyState } from '@/shared/ui/empty-state'
import { ErrorBanner } from '@/shared/ui/error-banner'
import { PageHeader } from '@/shared/ui/page-header'
import { Skeleton } from '@/shared/ui/skeleton'

export function AccessGrantCreatePage() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title="Выдать доступ"
        description="Грант действует только в пределах ваших собственных прав и потолка роли субъекта"
        breadcrumbs={[{ label: 'Доступы', to: '/access/grants' }, { label: 'Новый грант' }]}
      />
      <GrantForm
        initialSubject={subjectFromParams(searchParams)}
        onSaved={(grant: AccessGrant) => navigate(`/access/grants/${grant.id}`)}
        onCancel={() => navigate('/access/grants')}
      />
    </div>
  )
}

export function AccessGrantEditPage() {
  const { id = '' } = useParams()
  const navigate = useNavigate()
  const directory = useSubjectDirectory()
  const { data: grant, error, isLoading } = useAccessGrant(id)

  if (isLoading) return <Skeleton className="h-40 w-full" />
  if (error) return <ErrorBanner error={error} />
  if (!grant) return <EmptyState title="Грант не найден" />

  const subjectName = directory.subjectLabel(grant)
  const cardPath = `/access/grants/${grant.id}`

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title="Изменить грант"
        description={subjectName}
        breadcrumbs={[
          { label: 'Доступы', to: '/access/grants' },
          { label: subjectName, to: cardPath },
          { label: 'Изменение' },
        ]}
      />
      {isGrantEditable(grant) ? (
        <GrantForm
          grant={grant}
          onSaved={() => navigate(cardPath)}
          onCancel={() => navigate(cardPath)}
        />
      ) : (
        <EmptyState
          title="Грант нельзя изменить"
          description="Отозванный или истёкший грант не меняется. Выдайте новый грант."
        />
      )}
    </div>
  )
}
