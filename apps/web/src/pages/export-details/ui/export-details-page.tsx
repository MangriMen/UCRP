import type { ReactNode } from 'react'
import { Link, useParams } from 'react-router'
import {
  DownloadExportButton,
  ExportStatusBadge,
  exportChannelLabels,
  exportFormatLabels,
  useExport,
  type Export,
} from '@/entities/export'
import { useAuth } from '@/entities/user'
import { formatDateTime, formatNumber } from '@/shared/lib/format'
import { useUrlTab } from '@/shared/lib/hooks/use-url-tab'
import { Alert, AlertDescription } from '@/shared/ui/alert'
import { Card, CardContent } from '@/shared/ui/card'
import { EmptyState } from '@/shared/ui/empty-state'
import { ErrorBanner } from '@/shared/ui/error-banner'
import { PageHeader } from '@/shared/ui/page-header'
import { Skeleton } from '@/shared/ui/skeleton'
import { StatCard, StatGrid } from '@/shared/ui/stat-card'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/shared/ui/tabs'
import { MembersTable } from './members-table'

const TABS = ['recipients', 'control'] as const

const count = (value: number | null | undefined) => (value == null ? 'скрыто' : formatNumber(value))

function Row({ label, children }: { label: string; children: ReactNode }) {
  return (
    <div className="flex flex-col gap-1 sm:flex-row sm:gap-4">
      <dt className="text-sm text-muted-foreground sm:w-48">{label}</dt>
      <dd className="text-sm">{children}</dd>
    </div>
  )
}

function ExportCounters({ item }: { item: Export }) {
  return (
    <StatGrid className="xl:grid-cols-6">
      <StatCard label="В сегменте" value={count(item.totalInSegment)} />
      <StatCard label="С контактом" value={count(item.withContactCount)} />
      <StatCard label="Без контакта" value={count(item.noContactCount)} />
      <StatCard label="Исключено лимитом" value={count(item.blockedByLimitCount)} />
      <StatCard label="Контрольная группа" value={count(item.controlCount)} />
      <StatCard label="Получателей" value={count(item.exportedCount)} />
    </StatGrid>
  )
}

export function ExportDetailsPage() {
  const { id = '' } = useParams()
  const { can } = useAuth()
  const [tab, setTab] = useUrlTab(TABS, 'recipients')
  const { data: item, error, isLoading } = useExport(id)

  if (isLoading) return <Skeleton className="h-40 w-full" />
  if (error) return <ErrorBanner error={error} />
  if (!item) return <EmptyState title="Выгрузка не найдена" />

  const segmentName = item.segmentName ?? item.segmentId
  const title = `Выгрузка «${segmentName}»`
  const isReady = item.status === 'ready'

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title={title}
        breadcrumbs={[{ label: 'Выгрузки', to: '/exports' }, { label: title }]}
        description={`${exportChannelLabels[item.channel]} · ${formatDateTime(item.createdAt)}`}
        actions={
          <>
            <ExportStatusBadge status={item.status} />
            {isReady && can('export.download') && (
              <DownloadExportButton exportId={item.id} format={item.format} />
            )}
          </>
        }
      />
      {item.status === 'failed' && (
        <Alert variant="destructive">
          <AlertDescription>Выгрузку не удалось сформировать. Создайте её заново.</AlertDescription>
        </Alert>
      )}
      {item.suppressed && (
        <Alert>
          <AlertDescription>
            Счётчики скрыты: одно из значений меньше 5, по нему можно определить отдельного гостя.
          </AlertDescription>
        </Alert>
      )}
      <ExportCounters item={item} />
      <Card>
        <CardContent>
          <dl className="flex flex-col gap-3">
            <Row label="Сегмент">
              {can('segment.read') ? (
                <Link className="underline" to={`/segments/${item.segmentId}`}>
                  {segmentName}
                </Link>
              ) : (
                segmentName
              )}
              {item.segmentVersionNo != null && ` · версия ${item.segmentVersionNo}`}
            </Row>
            {item.campaignId && (
              <Row label="Кампания">
                {can('campaign.read') ? (
                  <Link className="underline" to={`/campaigns/${item.campaignId}`}>
                    Открыть кампанию
                  </Link>
                ) : (
                  item.campaignId
                )}
              </Row>
            )}
            <Row label="Канал">{exportChannelLabels[item.channel]}</Row>
            <Row label="Формат">{exportFormatLabels[item.format]}</Row>
            <Row label="Контрольная группа">{item.controlPercent ?? 0}%</Row>
            <Row label="Создана">{formatDateTime(item.createdAt)}</Row>
            <Row label="Готова">{item.finishedAt ? formatDateTime(item.finishedAt) : '—'}</Row>
          </dl>
        </CardContent>
      </Card>
      {can('export.members.read') && isReady && (
        <Tabs value={tab} onValueChange={setTab}>
          <TabsList>
            <TabsTrigger value="recipients">Получатели</TabsTrigger>
            <TabsTrigger value="control">Контрольная группа</TabsTrigger>
          </TabsList>
          <TabsContent value="recipients" className="pt-4">
            <MembersTable exportId={item.id} isControl={false} />
          </TabsContent>
          <TabsContent value="control" className="pt-4">
            <MembersTable exportId={item.id} isControl />
          </TabsContent>
        </Tabs>
      )}
    </div>
  )
}
