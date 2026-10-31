export const ALL_OPTION = 'all'

export const optionParam = (value: string) => (value === ALL_OPTION ? null : value)

export const parsePage = (value: string | null) => {
  const page = Number(value)
  return Number.isInteger(page) && page > 0 ? page : 1
}
