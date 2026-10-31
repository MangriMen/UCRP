import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { http, HttpResponse } from 'msw'
import { describe, expect, it } from 'vitest'
import { server } from '@/app/mocks/server'
import { API, apiError } from '@/shared/api/mocks/helpers'
import { SyncSourceButton } from './sync-source-button'

function renderButton(isRunning: boolean) {
  render(
    <QueryClientProvider
      client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
    >
      <SyncSourceButton sourceId="ds-0001" isRunning={isRunning} />
    </QueryClientProvider>,
  )
}

describe('SyncSourceButton', () => {
  it('is disabled while an import is running', () => {
    renderButton(true)
    expect(screen.getByRole('button', { name: 'Синхронизация…' })).toBeDisabled()
  })

  it('requests a sync of the given source', async () => {
    let syncedId: string | readonly string[] | undefined
    server.use(
      http.post(`${API}/data-sources/:id/sync`, ({ params }) => {
        syncedId = params.id
        return HttpResponse.json({ batchId: 'batch-x', status: 'queued' }, { status: 202 })
      }),
    )
    renderButton(false)
    await userEvent.click(screen.getByRole('button', { name: 'Синхронизировать' }))
    await waitFor(() => expect(syncedId).toBe('ds-0001'))
  })

  it('shows the conflict returned by the server', async () => {
    server.use(
      http.post(`${API}/data-sources/:id/sync`, () =>
        apiError(409, 'sync_in_progress', 'Синхронизация уже выполняется'),
      ),
    )
    renderButton(false)
    await userEvent.click(screen.getByRole('button', { name: 'Синхронизировать' }))
    expect(await screen.findByText('Синхронизация уже выполняется')).toBeInTheDocument()
  })
})
