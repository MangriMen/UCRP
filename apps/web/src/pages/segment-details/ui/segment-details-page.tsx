import { useParams } from 'react-router'
import { isSegmentComputing, RuleSummary, SegmentStatusBadge, useSegment } from '@/entities/segment'
import { useAuth } from '@/entities/user'
import { BuildSegmentButton } from '@/features/build-segment'
import { CreateExportButton } from '@/features/create-export'
import { RecomputeSegmentButton } from '@/features/recompute-segment'
import { formatDateTime } from '@/shared/lib/format'
import { Alert, AlertDescription } from '@/shared/ui/alert'
import { Card, CardContent, CardHeader, CardTitle } from '@/shared/ui/card'
import { EmptyState } from '@/shared/ui/empty-state'
import { ErrorBanner } from '@/shared/ui/error-banner'
import { PageHeader } from '@/shared/ui/page-header'
import { Skeleton } from '@/shared/ui/skeleton'
import { MemberTable } from './member-table'
import { SegmentRecommendations } from './segment-recommendations'
import { VersionHistory } from './version-history'

export function SegmentDetailsPage() {
  const { id = '' } = useParams()
  const { can } = useAuth()
  const { data: segment, error, isLoading } = useSegment(id)

  if (isLoading) return <Skeleton className="h-40 w-full" />
  if (error) return <ErrorBanner error={error} />
  if (!segment) return <EmptyState title="Сегмент не найден" />

  const canManage = can('segment.write')

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title={segment.name}
        breadcrumbs={[{ label: 'Сегменты', to: '/segments' }, { label: segment.name }]}
        description={segment.description ?? undefined}
        actions={
          <>
            <SegmentStatusBadge status={segment.status} />
            {canManage && <BuildSegmentButton segment={segment} />}
            {canManage && <RecomputeSegmentButton segmentId={segment.id} status={segment.status} />}
            {can('export.create') && (
              <CreateExportButton segmentId={segment.id} disabled={segment.status !== 'ready'} />
            )}
          </>
        }
      />
      {segment.status === 'failed' && (
        <Alert variant="destructive">
          <AlertDescription>
            Не удалось рассчитать сегмент. Запустите пересчёт ещё раз.
          </AlertDescription>
        </Alert>
      )}
      <div className="grid grid-cols-1 gap-4 lg:grid-cols-3">
        <Card>
          <CardHeader>
            <CardTitle className="text-sm font-normal text-muted-foreground">Аудитория</CardTitle>
          </CardHeader>
          <CardContent className="text-2xl font-semibold">{segment.memberCount}</CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle className="text-sm font-normal text-muted-foreground">Версия</CardTitle>
          </CardHeader>
          <CardContent className="text-2xl font-semibold">v{segment.currentVersionNo}</CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle className="text-sm font-normal text-muted-foreground">Рассчитан</CardTitle>
          </CardHeader>
          <CardContent className="text-2xl font-semibold">
            {segment.computedAt ? formatDateTime(segment.computedAt) : '—'}
          </CardContent>
        </Card>
      </div>
      {(segment.rule || segment.ruleText) && (
        <Card>
          <CardHeader>
            <CardTitle>Правило</CardTitle>
          </CardHeader>
          <CardContent className="text-sm">
            {segment.ruleText ? (
              <p>{segment.ruleText}</p>
            ) : segment.rule ? (
              <RuleSummary rule={segment.rule} />
            ) : null}
          </CardContent>
        </Card>
      )}
      {can('recommendation.read') && (
        <SegmentRecommendations
          segmentId={segment.id}
          isSegmentComputing={isSegmentComputing(segment.status)}
        />
      )}
      {canManage && (
        <>
          <section className="flex flex-col gap-3">
            <h2 className="text-lg font-semibold">Участники</h2>
            <MemberTable key={segment.id} segmentId={segment.id} />
          </section>
          <section className="flex flex-col gap-3">
            <h2 className="text-lg font-semibold">История версий</h2>
            <VersionHistory segmentId={segment.id} />
          </section>
        </>
      )}
    </div>
  )
}
