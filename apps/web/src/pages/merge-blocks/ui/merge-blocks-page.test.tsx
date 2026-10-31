import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import { describe, expect, it } from 'vitest'
import { MergeBlocksPage } from './merge-blocks-page'

describe('MergeBlocksPage', () => {
  it('lists blocks with a reason and an unblock action', async () => {
    render(
      <QueryClientProvider
        client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
      >
        <MemoryRouter>
          <MergeBlocksPage />
        </MemoryRouter>
      </QueryClientProvider>,
    )
    expect(await screen.findByText('Отклонённое объединение')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Снять запрет' })).toBeInTheDocument()
  })
})
