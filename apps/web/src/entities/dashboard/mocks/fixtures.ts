import type { DashboardMetrics, CrossTenantPoint } from '../model/types'

export const mockMetrics: Omit<DashboardMetrics, 'from' | 'to'> = {
  matchRate: 0.74,
  disputedMergeRate: 0.014,
  segmentToExportMedianSeconds: 185,
  segmentToExportP95Seconds: 410,
  conversionRate: 0.032,
  conversionLift: 0.011,
  sampleSizes: { events: 48200, guests: 9150, mergeCandidates: 320, exports: 18, contacts: 5400 },
}

export const mockCrossTenantPoints: Omit<CrossTenantPoint, 'period'>[] = [
  { one: 5200, two: 2100, threePlus: 900, crossTenantRate: 0.366 },
  { one: 5000, two: 2200, threePlus: 1000, crossTenantRate: 0.39 },
  { one: 4800, two: 2300, threePlus: 1150, crossTenantRate: 0.418 },
  { one: 4700, two: 2350, threePlus: 1250, crossTenantRate: 0.434 },
]
