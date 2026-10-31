export const SEGMENT_TOP_TENANTS_LIMITS = [10, 20, 50] as const
export const SEGMENT_TOP_TENANTS_LIMIT = SEGMENT_TOP_TENANTS_LIMITS[0]
export const SEGMENT_TOP_TENANTS_SUPPRESSION_THRESHOLD = 5

export const parseSegmentTopTenantsLimit = (value: string | null) => {
  const limit = Number(value)
  return SEGMENT_TOP_TENANTS_LIMITS.find((option) => option === limit) ?? SEGMENT_TOP_TENANTS_LIMIT
}

export const recommendationKeys = {
  all: ['recommendations'] as const,
  guest: (guestId: string) => [...recommendationKeys.all, 'guest', guestId] as const,
  segment: (segmentId: string, limit: number) =>
    [...recommendationKeys.all, 'segment', segmentId, limit] as const,
}
