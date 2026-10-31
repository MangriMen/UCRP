const dateTime = new Intl.DateTimeFormat('ru-RU', { dateStyle: 'short', timeStyle: 'short' })
const date = new Intl.DateTimeFormat('ru-RU', { dateStyle: 'medium' })
const number = new Intl.NumberFormat('ru-RU')
const time = new Intl.DateTimeFormat('ru-RU', { timeStyle: 'short' })
const rubles = new Intl.NumberFormat('ru-RU', { style: 'currency', currency: 'RUB' })

export const formatDateTime = (iso: string) => dateTime.format(new Date(iso))
export const formatDate = (iso: string) => date.format(new Date(iso))
export const formatNumber = (value: number) => number.format(value)
export const formatTime = (iso: string) => time.format(new Date(iso))
export const formatMoney = (amountMinor: number) => rubles.format(amountMinor / 100)

const percent = new Intl.NumberFormat('ru-RU', { style: 'percent', maximumFractionDigits: 1 })

export const formatPercent = (ratio: number) => percent.format(ratio)
