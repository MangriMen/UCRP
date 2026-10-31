import { useSearchParams } from 'react-router'
import { parseSegmentTopTenantsLimit, SegmentTopTenants } from '@/entities/recommendation'
import { useTenantCategories } from '@/entities/tenant-category'
import { applyParamChanges } from '@/shared/lib/list-params'

const LIMIT_PARAM = 'top'

interface SegmentRecommendationsProps {
  segmentId: string
  isSegmentComputing: boolean
}

export function SegmentRecommendations({
  segmentId,
  isSegmentComputing,
}: SegmentRecommendationsProps) {
  const [searchParams, setSearchParams] = useSearchParams()
  const categories = useTenantCategories()
  const limit = parseSegmentTopTenantsLimit(searchParams.get(LIMIT_PARAM))

  return (
    <SegmentTopTenants
      segmentId={segmentId}
      isSegmentComputing={isSegmentComputing}
      limit={limit}
      onLimitChange={(next) =>
        setSearchParams(
          (current) => applyParamChanges(current, { [LIMIT_PARAM]: String(next) }, false),
          { replace: true },
        )
      }
      categoryName={(categoryId) => categories.data?.find((item) => item.id === categoryId)?.name}
    />
  )
}
