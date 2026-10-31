export const SEGMENT_TOP_TENANTS_LIMIT = 10

export const recommendationKeys = {
  all: ['recommendations'] as const,
  guest: (guestId: string) => [...recommendationKeys.all, 'guest', guestId] as const,
  segment: (segmentId: string, limit: number) =>
    [...recommendationKeys.all, 'segment', segmentId, limit] as const,
}
