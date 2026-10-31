import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it } from 'vitest'
import { ConfirmMergeButton } from './confirm-merge-button'

function renderButton(candidateId: string) {
  render(
    <QueryClientProvider
      client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
    >
      <ConfirmMergeButton candidateId={candidateId} />
    </QueryClientProvider>,
  )
}

describe('ConfirmMergeButton', () => {
  it('shows an explicit message when the candidate is already resolved', async () => {
    renderButton('candidate-0002')
    await userEvent.click(screen.getByRole('button', { name: 'Объединить' }))
    await userEvent.click(
      within(await screen.findByRole('alertdialog')).getByRole('button', { name: 'Объединить' }),
    )
    expect(await screen.findByText(/уже обработана/)).toBeInTheDocument()
  })

  it('stays quiet when the merge succeeds', async () => {
    renderButton('candidate-0001')
    await userEvent.click(screen.getByRole('button', { name: 'Объединить' }))
    await userEvent.click(
      within(await screen.findByRole('alertdialog')).getByRole('button', { name: 'Объединить' }),
    )
    await waitFor(() => expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument())
    expect(screen.queryByText('Конфликт')).not.toBeInTheDocument()
  })
})
