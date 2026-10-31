import { describe, expect, it } from 'vitest'
import type { Me } from '@/entities/user'
import { mockUsers } from '@/entities/user/mocks'
import { navigationFor } from './navigation'

const userWithRole = (role: string) => {
  const user = mockUsers.find((candidate) => candidate.role === role)
  if (!user) throw new Error(`no mock user for ${role}`)
  return user
}

const titles = (role: string) =>
  navigationFor(userWithRole(role)).flatMap((group) => group.items.map((item) => item.title))

const staffWithEveryPermission: Me = {
  ...userWithRole('manager'),
  permissions: [
    ...new Set(
      mockUsers.filter((user) => user.role !== 'tenant').flatMap((user) => user.permissions),
    ),
  ],
}

describe('navigationFor', () => {
  it('groups staff sections in the agreed order', () => {
    const layout = navigationFor(staffWithEveryPermission).map((group) => [
      group.label,
      group.items.map((item) => item.title),
    ])
    expect(layout).toEqual([
      ['Обзор', ['Дашборд', 'Качество данных']],
      ['Арендаторы', ['Арендаторы', 'Источники данных', 'Импорт данных']],
      ['Гости', ['Гости', 'Сегменты', 'Объединение профилей']],
      ['Коммуникации', ['Кампании', 'Выгрузки', 'Лимиты контактов']],
      [
        'Администрирование',
        ['Пользователи', 'Справочники', 'Запреты объединения', 'Журнал аудита'],
      ],
    ])
  })

  it('gives the tenant a single entry', () => {
    expect(titles('tenant')).toEqual(['Мой арендатор'])
  })

  it('gives a user without section permissions no entries', () => {
    expect(navigationFor({ ...userWithRole('manager'), role: 'system', permissions: [] })).toEqual(
      [],
    )
  })

  it('shows user administration to the manager only', () => {
    expect(titles('manager')).toContain('Пользователи')
    expect(titles('analyst')).not.toContain('Пользователи')
    expect(titles('marketer')).not.toContain('Пользователи')
  })

  it('shows guests to everyone who holds guest search', () => {
    expect(titles('manager')).toContain('Гости')
    expect(titles('marketer')).toContain('Гости')
    expect(titles('tenant')).not.toContain('Гости')
  })

  it('hides staff entries from the tenant', () => {
    expect(titles('tenant')).not.toContain('Выгрузки')
    expect(titles('tenant')).not.toContain('Справочники')
    expect(titles('tenant')).not.toContain('Сегменты')
  })

  it('hides the tenant entry from staff', () => {
    expect(titles('manager')).not.toContain('Мой арендатор')
  })
})
