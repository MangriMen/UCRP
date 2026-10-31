import { useNavigate, useSearchParams } from 'react-router'
import {
  CampaignStatusBadge,
  campaignStatusLabels,
  useCampaigns,
  type Campaign,
  type CampaignStatus,
} from '@/entities/campaign'
import { formatDate } from '@/shared/lib/format'
import { ALL_OPTION, parsePage } from '@/shared/lib/search-params'
import { DataTable, type Column } from '@/shared/ui/data-table'
import { OptionSelect } from '@/shared/ui/option-select'
import { PageHeader } from '@/shared/ui/page-header'

const columns: Column<Campaign>[] = [
  { id: 'name', header: 'Кампания', cell: (c) => c.name },
  { id: 'tenant', header: 'Арендатор', cell: (c) => c.targetTenantName ?? '—' },
  { id: 'segment', header: 'Сегмент', cell: (c) => c.segmentName ?? '—' },
  { id: 'status', header: 'Статус', cell: (c) => <CampaignStatusBadge status={c.status} /> },
  {
    id: 'period',
    header: 'Период',
    cell: (c) =>
      c.startsAt ? `${formatDate(c.startsAt)} — ${c.endsAt ? formatDate(c.endsAt) : '…'}` : '—',
  },
]

const statusItems = [
  { value: ALL_OPTION, label: 'Все статусы' },
  ...Object.entries(campaignStatusLabels).map(([value, label]) => ({ value, label })),
]

const isCampaignStatus = (value: string): value is CampaignStatus => value in campaignStatusLabels

export function CampaignsPage() {
  const navigate = useNavigate()
  const [searchParams, setSearchParams] = useSearchParams()
  const status = searchParams.get('status') ?? ALL_OPTION
  const page = parsePage(searchParams.get('page'))

  const updateParams = (changes: Record<string, string | null>) =>
    setSearchParams((current) => {
      const next = new URLSearchParams(current)
      for (const [key, value] of Object.entries(changes)) {
        if (value === null || value === ALL_OPTION) next.delete(key)
        else next.set(key, value)
      }
      return next
    })

  const campaigns = useCampaigns({
    status: isCampaignStatus(status) ? status : undefined,
    page,
  })

  return (
    <div className="flex flex-col gap-6">
      <PageHeader title="Кампании" description="Рассылки по сегментам и их эффективность" />
      <OptionSelect
        items={statusItems}
        value={status}
        onValueChange={(value) => updateParams({ status: value, page: null })}
        ariaLabel="Статус"
        className="w-56"
      />
      <DataTable
        columns={columns}
        rows={campaigns.data?.items}
        getRowId={(c) => c.id}
        isLoading={campaigns.isLoading}
        error={campaigns.error}
        emptyTitle="Кампаний пока нет"
        meta={campaigns.data?.meta}
        onPageChange={(next) => updateParams({ page: String(next) })}
        onRowClick={(c) => navigate(`/campaigns/${c.id}`)}
      />
    </div>
  )
}
