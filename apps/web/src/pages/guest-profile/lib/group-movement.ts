import type { GuestMovementStep } from '@/entities/guest'
import { dayKey } from './day-key'

export interface MovementLeg {
  key: string
  mallId: string
  mallName: string
  floor: number
  startedAt: string
  endedAt: string
  tenantNames: string[]
  purchaseCount: number
  stepCount: number
}

const continuesLeg = (leg: MovementLeg, step: GuestMovementStep) =>
  leg.mallId === step.mallId &&
  leg.floor === step.floor &&
  dayKey(leg.endedAt) === dayKey(step.occurredAt)

const startLeg = (step: GuestMovementStep): MovementLeg => ({
  key: step.eventId,
  mallId: step.mallId,
  mallName: step.mallName,
  floor: step.floor,
  startedAt: step.occurredAt,
  endedAt: step.occurredAt,
  tenantNames: [step.tenantName],
  purchaseCount: step.eventTypeCode === 'purchase' ? 1 : 0,
  stepCount: 1,
})

const extendLeg = (leg: MovementLeg, step: GuestMovementStep) => {
  leg.endedAt = step.occurredAt
  leg.stepCount += 1
  if (step.eventTypeCode === 'purchase') leg.purchaseCount += 1
  if (!leg.tenantNames.includes(step.tenantName)) leg.tenantNames.push(step.tenantName)
}

export function groupMovement(steps: GuestMovementStep[]): MovementLeg[] {
  const sorted = [...steps].sort((a, b) => a.occurredAt.localeCompare(b.occurredAt))
  const legs: MovementLeg[] = []
  for (const step of sorted) {
    const last = legs.at(-1)
    if (last && continuesLeg(last, step)) {
      extendLeg(last, step)
    } else {
      legs.push(startLeg(step))
    }
  }
  return legs
}
