export interface Period {
  from: string
  to: string
}

const dayOf = (date: Date) => date.toISOString().slice(0, 10)

export function defaultPeriod(months = 6, now = new Date()): Period {
  const year = now.getUTCFullYear()
  const month = now.getUTCMonth() - months
  const lastDay = new Date(Date.UTC(year, month + 1, 0)).getUTCDate()
  const start = new Date(Date.UTC(year, month, Math.min(now.getUTCDate(), lastDay)))
  return { from: dayOf(start), to: dayOf(now) }
}

export const isPeriodValid = ({ from, to }: Period) => from !== '' && to !== '' && from <= to

export const startOfDay = (day: string) => `${day}T00:00:00Z`
export const endOfDay = (day: string) => `${day}T23:59:59Z`

export const periodQuery = ({ from, to }: Period) => ({
  from: startOfDay(from),
  to: endOfDay(to),
})
