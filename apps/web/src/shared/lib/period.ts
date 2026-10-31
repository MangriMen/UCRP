export interface Period {
  from: string
  to: string
}

const dayOf = (date: Date) => date.toISOString().slice(0, 10)

export function defaultPeriod(months = 6, now = new Date()): Period {
  const start = new Date(now)
  start.setUTCMonth(start.getUTCMonth() - months)
  return { from: dayOf(start), to: dayOf(now) }
}

export const isPeriodValid = ({ from, to }: Period) => from !== '' && to !== '' && from <= to

export const periodQuery = ({ from, to }: Period) => ({
  from: `${from}T00:00:00Z`,
  to: `${to}T23:59:59Z`,
})
