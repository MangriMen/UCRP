import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { http, HttpResponse } from 'msw'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { server } from '@/app/mocks/server'
import { mockUsers } from '@/entities/user/mocks'
import { setAccessToken } from '@/shared/api/client'
import { API } from '@/shared/api/mocks/helpers'
import { MergeGuestsDialog } from './merge-guests-dialog'

const SESSION_KEY = 'ucrp-mock-session'
const [marketer] = mockUsers

const mergeResult = {
  winnerGuestId: 'guest-0002',
  mergedGuestId: 'guest-0001',
  eventsMoved: 14,
  identifiersMoved: 2,
}

async function openDialog(onMerged = vi.fn()) {
  render(
    <QueryClientProvider
      client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
    >
      <MergeGuestsDialog guestId="guest-0001" onMerged={onMerged} />
    </QueryClientProvider>,
  )
  await userEvent.click(screen.getByRole('button', { name: 'Объединить с другим профилем' }))
  return onMerged
}

describe('MergeGuestsDialog', () => {
  beforeEach(() => {
    localStorage.setItem(SESSION_KEY, marketer.userId)
  })

  afterEach(() => {
    localStorage.clear()
    setAccessToken(null)
  })

  it('lists found guests except the current one', async () => {
    await openDialog()
    await userEvent.type(screen.getByLabelText('Поиск гостя'), 'Ива')
    expect(await screen.findByText('Мария Иванова')).toBeInTheDocument()
    expect(screen.queryByText('Иван Петров')).not.toBeInTheDocument()
  })

  it('asks for a longer query instead of searching', async () => {
    await openDialog()
    await userEvent.type(screen.getByLabelText('Поиск гостя'), 'И')
    expect(screen.getByText('Введите не меньше 2 символов')).toBeInTheDocument()
    expect(screen.queryByText('Мария Иванова')).not.toBeInTheDocument()
  })

  it('keeps the submit button disabled until a guest is selected', async () => {
    await openDialog()
    expect(screen.getByRole('button', { name: 'Объединить' })).toBeDisabled()
    await userEvent.type(screen.getByLabelText('Поиск гостя'), 'Ива')
    const option = await screen.findByRole('button', { name: /Мария Иванова/ })
    await userEvent.click(option)
    expect(option).toHaveAttribute('aria-pressed', 'true')
    expect(screen.getByRole('button', { name: 'Объединить' })).toBeEnabled()
  })

  it('merges into the selected guest', async () => {
    let requestedTarget: unknown
    server.use(
      http.post(`${API}/guests/:id/merge`, async ({ request }) => {
        requestedTarget = ((await request.json()) as { targetGuestId: string }).targetGuestId
        return HttpResponse.json(mergeResult)
      }),
    )
    const onMerged = await openDialog()
    await userEvent.type(screen.getByLabelText('Поиск гостя'), 'Ива')
    await userEvent.click(await screen.findByRole('button', { name: /Мария Иванова/ }))
    await userEvent.click(screen.getByRole('button', { name: 'Объединить' }))
    await waitFor(() => expect(onMerged).toHaveBeenCalledWith(mergeResult))
    expect(requestedTarget).toBe('guest-0002')
  })
})
