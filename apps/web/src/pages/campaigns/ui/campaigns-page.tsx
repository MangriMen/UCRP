import { useNavigate } from 'react-router'
import {
  CampaignStatusBadge,
  campaignStatusLabels,
  useCampaigns,
  type Campaign,
  type CampaignStatus,
} from '@/entities/campaign'
import { formatDate } from '@/shared/lib/format'
import { useOptionParam } from '@/shared/lib/hooks/use-option-param'
import { useListParams } from '@/shared/lib/list-params'
import { ALL_OPTION } from '@/shared/lib/search-params'
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

const statusValues = statusItems.map((item) => item.value)

export function CampaignsPage() {
  const navigate = useNavigate()
  const params = useListParams()
  const { page, pageSize, setPage, setPageSize } = params
  const statusParam = useOptionParam(params, 'status', statusValues)

  const campaigns = useCampaigns({
    status: statusParam.value === ALL_OPTION ? undefined : (statusParam.value as CampaignStatus),
    page,
    pageSize,
  })

  return (
    <div className="flex flex-col gap-6">
      <PageHeader title="Кампании" description="Рассылки по сегментам и их эффективность" />
      <OptionSelect
        items={statusItems}
        value={statusParam.value}
        onValueChange={statusParam.onChange}
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
        onPageChange={setPage}
        onPageSizeChange={setPageSize}
        onRowClick={(c) => navigate(`/campaigns/${c.id}`)}
      />
    </div>
  )
}
