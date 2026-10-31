import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { UnmergeBlocksPage } from './unmerge-blocks-page'

describe('UnmergeBlocksPage', () => {
  it('lists blocks with a reason and an unblock action', async () => {
    render(
      <QueryClientProvider
        client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
      >
        <UnmergeBlocksPage />
      </QueryClientProvider>,
    )
    expect(await screen.findByText('Отклонённая склейка')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Снять запрет' })).toBeInTheDocument()
  })
})
