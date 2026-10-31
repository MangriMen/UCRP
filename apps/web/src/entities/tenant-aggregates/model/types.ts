import type { Schema } from '@/shared/api/types'

export type TenantVisitAggregate = Schema<'ManagementTenantAggregate'>
export type TenantAggregateDay = Schema<'TenantAggregateDay'>
export type EmptyTenantAggregate = Schema<'EmptyTenantAggregate'>
export type TenantAggregate = TenantVisitAggregate | EmptyTenantAggregate

export const isVisitAggregate = (value: TenantAggregate): value is TenantVisitAggregate =>
  'visitsCount' in value
