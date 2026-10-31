import { useNavigate, useSearchParams } from 'react-router'
import {
  CampaignStatusBadge,
  campaignStatusLabels,
  useCampaigns,
  type Campaign,
  type CampaignStatus,
} from '@/entities/campaign'
import { formatDate } from '@/shared/lib/format'
import { DataTable, type Column } from '@/shared/ui/data-table'
import { PageHeader } from '@/shared/ui/page-header'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/shared/ui/select'

const ALL = 'all'

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
  { value: ALL, label: 'Все статусы' },
  ...Object.entries(campaignStatusLabels).map(([value, label]) => ({ value, label })),
]

const isCampaignStatus = (value: string): value is CampaignStatus => value in campaignStatusLabels

const parsePage = (value: string | null) => {
  const page = Number(value)
  return Number.isInteger(page) && page > 0 ? page : 1
}

export function CampaignsPage() {
  const navigate = useNavigate()
  const [searchParams, setSearchParams] = useSearchParams()
  const status = searchParams.get('status') ?? ALL
  const page = parsePage(searchParams.get('page'))

  const updateParams = (changes: Record<string, string | null>) =>
    setSearchParams((current) => {
      const next = new URLSearchParams(current)
      for (const [key, value] of Object.entries(changes)) {
        if (value === null || value === ALL) next.delete(key)
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
      <Select
        items={statusItems}
        value={status}
        onValueChange={(value) => updateParams({ status: value, page: null })}
      >
        <SelectTrigger className="w-56" aria-label="Статус">
          <SelectValue />
        </SelectTrigger>
        <SelectContent>
          {statusItems.map((item) => (
            <SelectItem key={item.value} value={item.value}>
              {item.label}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
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
