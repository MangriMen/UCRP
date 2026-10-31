import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router'
import { afterEach, describe, expect, it } from 'vitest'
import { RoleGuard } from '@/entities/user'
import { setAccessToken } from '@/shared/api/client'
import { ContactLimitsPage } from './contact-limits-page'

function renderAs(userId: string) {
  localStorage.setItem('ucrp-mock-session', userId)
  render(
    <QueryClientProvider
      client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
    >
      <MemoryRouter initialEntries={['/settings/contact-limits']}>
        <Routes>
          <Route
            path="/settings/contact-limits"
            element={
              <RoleGuard>
                <ContactLimitsPage />
              </RoleGuard>
            }
          />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

afterEach(() => {
  localStorage.clear()
  setAccessToken(null)
})

describe('ContactLimitsPage', () => {
  it.each([
    ['user-marketer', true],
    ['user-analyst', false],
    ['user-manager', false],
  ])('%s can edit limits: %s', async (userId, allowed) => {
    renderAs(userId)
    await screen.findByText('Лимиты контактов')
    expect(await screen.findByRole('table')).toBeInTheDocument()
    expect(!!screen.queryByText('Новый лимит')).toBe(allowed)
  })
})
