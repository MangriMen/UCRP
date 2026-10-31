const numberFormat = new Intl.NumberFormat('ru-RU')

export const formatAggregateCell = (value: number | null | undefined, threshold: number) =>
  value === null || value === undefined ? `меньше ${threshold}` : numberFormat.format(value)
