import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { http } from 'msw'
import { describe, expect, it, vi } from 'vitest'
import { server } from '@/app/mocks/server'
import { API, apiError } from '@/shared/api/mocks/helpers'
import { ExportWizard } from './export-wizard'

function renderWizard() {
  render(
    <QueryClientProvider
      client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
    >
      <ExportWizard segmentId="segment-0001" onClose={vi.fn()} />
    </QueryClientProvider>,
  )
}

async function toConfirm() {
  await userEvent.click(screen.getByRole('button', { name: 'Далее' }))
  expect(await screen.findByText('Будет отправлено')).toBeInTheDocument()
  await userEvent.click(screen.getByRole('button', { name: 'Далее' }))
}

describe('ExportWizard', () => {
  it('walks through the steps and shows the result', async () => {
    renderWizard()
    await toConfirm()
    await userEvent.click(screen.getByRole('button', { name: 'Создать выгрузку' }))
    expect(await screen.findByText(/Выгрузка создана/)).toBeInTheDocument()
  })

  it('explains the contact limit conflict', async () => {
    server.use(
      http.post(`${API}/exports`, () =>
        apiError(409, 'contact_limit_exceeded', 'no recipients left'),
      ),
    )
    renderWizard()
    await toConfirm()
    await userEvent.click(screen.getByRole('button', { name: 'Создать выгрузку' }))
    expect(await screen.findByText(/исключены лимитом контактов/)).toBeInTheDocument()
  })
})
