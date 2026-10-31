import { describe, expect, it } from 'vitest'
import type { GuestMovementStep } from '@/entities/guest'
import { groupMovement } from './group-movement'

const step = (
  eventId: string,
  occurredAt: string,
  overrides: Partial<GuestMovementStep> = {},
): GuestMovementStep => ({
  eventId,
  eventTypeCode: 'visit',
  tenantId: 'tenant-0001',
  tenantName: 'Лента',
  mallId: 'mall-0001',
  mallName: 'Центральный ТЦ',
  floor: 1,
  occurredAt,
  ...overrides,
})

describe('groupMovement', () => {
  it('returns no legs for no steps', () => {
    expect(groupMovement([])).toEqual([])
  })

  it('merges consecutive steps on one floor of one day into a leg', () => {
    const legs = groupMovement([
      step('a', '2026-10-20T09:00:00'),
      step('b', '2026-10-20T09:30:00', {
        tenantId: 'tenant-0009',
        tenantName: 'Зара',
        eventTypeCode: 'purchase',
      }),
    ])

    expect(legs).toHaveLength(1)
    expect(legs[0]).toMatchObject({
      startedAt: '2026-10-20T09:00:00',
      endedAt: '2026-10-20T09:30:00',
      tenantNames: ['Лента', 'Зара'],
      purchaseCount: 1,
      stepCount: 2,
    })
  })

  it('orders legs chronologically regardless of input order', () => {
    const legs = groupMovement([
      step('late', '2026-10-20T15:00:00', { floor: 3 }),
      step('early', '2026-10-20T09:00:00', { floor: 1 }),
    ])

    expect(legs.map((leg) => leg.floor)).toEqual([1, 3])
  })

  it('starts a new leg when the floor changes and returns later to the first floor', () => {
    const legs = groupMovement([
      step('a', '2026-10-20T09:00:00', { floor: 1 }),
      step('b', '2026-10-20T10:00:00', { floor: 2 }),
      step('c', '2026-10-20T11:00:00', { floor: 1 }),
    ])

    expect(legs.map((leg) => leg.floor)).toEqual([1, 2, 1])
  })

  it('treats the same floor number in another mall as a different zone', () => {
    const legs = groupMovement([
      step('a', '2026-10-20T09:00:00'),
      step('b', '2026-10-20T10:00:00', { mallId: 'mall-0002', mallName: 'Северный ТЦ' }),
    ])

    expect(legs.map((leg) => leg.mallId)).toEqual(['mall-0001', 'mall-0002'])
  })

  it('starts a new leg on another day even on the same floor', () => {
    const legs = groupMovement([step('a', '2026-10-20T09:00:00'), step('b', '2026-10-21T09:00:00')])

    expect(legs).toHaveLength(2)
  })
})
