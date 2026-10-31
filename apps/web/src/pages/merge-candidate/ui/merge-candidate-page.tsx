import { useParams } from 'react-router'
import {
  MergeResolutionBadge,
  useMergeCandidate,
  type MergeResolution,
} from '@/entities/merge-candidate'
import { useAuth } from '@/entities/user'
import { ConfirmMergeButton } from '@/features/confirm-merge'
import { RejectMergeButton } from '@/features/reject-merge'
import { RevertMergeButton } from '@/features/revert-merge'
import { EmptyState } from '@/shared/ui/empty-state'
import { ErrorBanner } from '@/shared/ui/error-banner'
import { PageHeader } from '@/shared/ui/page-header'
import { Skeleton } from '@/shared/ui/skeleton'
import { MergeCompare } from './merge-compare'

const REVERTIBLE: MergeResolution[] = ['confirmed', 'auto_merged']

export function MergeCandidatePage() {
  const { id = '' } = useParams()
  const { can } = useAuth()
  const { data: detail, error, isLoading } = useMergeCandidate(id)

  if (isLoading) return <Skeleton className="h-40 w-full" />
  if (error) return <ErrorBanner error={error} />
  if (!detail) return <EmptyState title="Кандидат на объединение не найден" />

  const canResolve = can('identity.decide')

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title="Кандидат на объединение"
        breadcrumbs={[
          { label: 'Объединение профилей', to: '/identity/merge-candidates' },
          { label: 'Кандидат на объединение' },
        ]}
        description={detail.reasonText ?? detail.reasonCode}
      />
      <div className="flex flex-wrap items-center gap-3 text-sm">
        <MergeResolutionBadge resolution={detail.resolution} />
        <span className="text-muted-foreground">Уверенность: {detail.score}%</span>
      </div>
      <MergeCompare detail={detail} />
      {canResolve && detail.resolution === 'pending' && (
        <div className="flex flex-wrap items-start gap-2">
          <ConfirmMergeButton candidateId={detail.id} />
          <RejectMergeButton candidateId={detail.id} />
        </div>
      )}
      {canResolve && REVERTIBLE.includes(detail.resolution) && (
        <div>
          <RevertMergeButton candidateId={detail.id} />
        </div>
      )}
    </div>
  )
}
