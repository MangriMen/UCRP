import { describe, expect, it } from 'vitest'
import { mockUsers } from '@/entities/user/mocks'
import { navigationFor } from './navigation'

const titles = (role: string) => {
  const user = mockUsers.find((candidate) => candidate.role === role)
  if (!user) throw new Error(`no mock user for ${role}`)
  return navigationFor(user).flatMap((group) => group.items.map((item) => item.title))
}

describe('navigationFor', () => {
  it('gives the tenant a single entry', () => {
    expect(titles('tenant')).toEqual(['Мой арендатор'])
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
