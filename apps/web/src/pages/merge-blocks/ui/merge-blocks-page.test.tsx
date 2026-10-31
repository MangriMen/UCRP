import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { MergeBlocksPage } from './merge-blocks-page'

describe('MergeBlocksPage', () => {
  it('lists blocks with a reason and an unblock action', async () => {
    render(
      <QueryClientProvider
        client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
      >
        <MergeBlocksPage />
      </QueryClientProvider>,
    )
    expect(await screen.findByText('Отклонённая склейка')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Снять запрет' })).toBeInTheDocument()
  })
})
