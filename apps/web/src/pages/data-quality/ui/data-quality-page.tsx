import { useSearchParams } from 'react-router'
import { useDataQualitySources } from '@/entities/data-quality'
import { PageHeader } from '@/shared/ui/page-header'
import { SourceDetail } from './source-detail'
import { SourceQualityTable } from './source-quality-table'

export function DataQualityPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const selected = searchParams.get('source')
  const sources = useDataQualitySources()

  const select = (id: string | null) => setSearchParams(id ? { source: id } : {}, { replace: true })

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title="Качество данных"
        description="Доля отклонённых строк, свежесть импорта и спорные профили по источникам"
      />
      <SourceQualityTable
        rows={sources.data}
        isLoading={sources.isLoading}
        error={sources.error}
        onSelect={select}
      />
      {selected && <SourceDetail sourceId={selected} onClose={() => select(null)} />}
    </div>
  )
}
