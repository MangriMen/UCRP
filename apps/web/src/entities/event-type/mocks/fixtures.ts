import type { EventType } from '../model/types'

export const EVENT_TYPE_IN_USE_ID = 'event-type-0001'

export const mockEventTypes: EventType[] = [
  {
    id: 'event-type-0001',
    code: 'purchase',
    name: 'Покупка',
    hasAmount: true,
    countsAsPresence: true,
    isActive: true,
  },
  {
    id: 'event-type-0002',
    code: 'wifi_connect',
    name: 'Подключение к Wi-Fi',
    hasAmount: false,
    countsAsPresence: true,
    isActive: true,
  },
  {
    id: 'event-type-0003',
    code: 'push_open',
    name: 'Открытие push',
    hasAmount: false,
    countsAsPresence: false,
    isActive: false,
  },
]
