import { describe, expect, it } from 'vitest'
import { navigationFor } from './navigation'

const titles = (role: Parameters<typeof navigationFor>[0]) =>
  navigationFor(role).flatMap((group) => group.items.map((item) => item.title))

describe('navigationFor', () => {
  it('gives the tenant a single entry', () => {
    expect(titles('tenant')).toEqual(['Мой арендатор'])
  })

  it('hides empty groups', () => {
    expect(navigationFor('tenant')).toHaveLength(1)
  })

  it('shows user administration to the manager only', () => {
    expect(titles('manager')).toContain('Пользователи')
    expect(titles('analyst')).not.toContain('Пользователи')
    expect(titles('marketer')).not.toContain('Пользователи')
  })

  it('lets only the analyst manage merge blocks', () => {
    expect(titles('analyst')).toContain('Запреты объединения')
    expect(titles('marketer')).toContain('Объединение профилей')
    expect(titles('marketer')).not.toContain('Запреты объединения')
  })

  it('hides segment and export write entries from the tenant', () => {
    expect(titles('tenant')).not.toContain('Выгрузки')
    expect(titles('tenant')).not.toContain('Лимиты контактов')
    expect(titles('tenant')).not.toContain('Сегменты')
  })
})
