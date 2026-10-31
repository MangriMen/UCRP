const MAX_CONTACTS_LIMIT = 100
const MAX_PERIOD_DAYS = 365

interface LimitErrors {
  maxContacts?: string
  periodDays?: string
}

const isIntegerInRange = (value: string, max: number) => {
  const number = Number(value)
  return value.trim() !== '' && Number.isInteger(number) && number >= 1 && number <= max
}

export function validateLimit(maxContacts: string, periodDays: string): LimitErrors {
  return {
    maxContacts: isIntegerInRange(maxContacts, MAX_CONTACTS_LIMIT)
      ? undefined
      : `Целое число от 1 до ${MAX_CONTACTS_LIMIT}`,
    periodDays: isIntegerInRange(periodDays, MAX_PERIOD_DAYS)
      ? undefined
      : `Целое число от 1 до ${MAX_PERIOD_DAYS}`,
  }
}
