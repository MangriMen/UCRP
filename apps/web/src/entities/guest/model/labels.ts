import type { ContactChannel, GuestStatus, IdentifierType } from './types'

export const guestStatusLabels: Record<GuestStatus, string> = {
  active: 'Активен',
  merged: 'Объединён',
  deleted: 'Удалён',
}

export const identifierTypeLabels: Record<IdentifierType, string> = {
  phone: 'Телефон',
  email: 'Email',
  card: 'Карта',
  loyalty_id: 'Карта лояльности',
  wifi_mac: 'Wi-Fi MAC',
  car_plate: 'Номер автомобиля',
}

export const contactChannelLabels: Record<ContactChannel, string> = {
  sms: 'SMS',
  push: 'Push',
  email: 'Email',
}
