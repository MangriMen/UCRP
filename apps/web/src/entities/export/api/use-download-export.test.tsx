import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { http } from 'msw'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { server } from '@/app/mocks/server'
import { setAccessToken } from '@/shared/api/client'
import { API, apiError } from '@/shared/api/mocks/helpers'
import * as saveBlobModule from '@/shared/lib/save-blob'
import { DownloadExportButton } from '../ui/download-export-button'

function renderButton(format: 'csv' | 'xlsx' = 'csv') {
  render(
    <QueryClientProvider client={new QueryClient()}>
      <DownloadExportButton exportId="export-0001" format={format} />
    </QueryClientProvider>,
  )
}

afterEach(() => {
  vi.restoreAllMocks()
  setAccessToken(null)
})

describe('DownloadExportButton', () => {
  it('fetches the file with the bearer token and saves it', async () => {
    setAccessToken('secret')
    let auth: string | null = null
    server.use(
      http.get(`${API}/exports/:id/download`, ({ request }) => {
        auth = request.headers.get('Authorization')
        return new Response('a,b\n', { headers: { 'Content-Type': 'text/csv' } })
      }),
    )
    const save = vi.spyOn(saveBlobModule, 'saveBlob').mockImplementation(() => {})
    renderButton()
    await userEvent.click(screen.getByRole('button', { name: 'Скачать CSV' }))
    await waitFor(() => expect(save).toHaveBeenCalledTimes(1))
    const [blob, filename] = save.mock.calls[0]
    expect(filename).toBe('export-export-0001.csv')
    expect(blob.type).toBe('text/csv')
    expect(blob.size).toBe(4)
    expect(auth).toBe('Bearer secret')
  })

  it('saves an Excel export with the xlsx extension', async () => {
    const save = vi.spyOn(saveBlobModule, 'saveBlob').mockImplementation(() => {})
    renderButton('xlsx')
    await userEvent.click(screen.getByRole('button', { name: 'Скачать Excel' }))
    await waitFor(() => expect(save).toHaveBeenCalledTimes(1))
    expect(save.mock.calls[0][1]).toBe('export-export-0001.xlsx')
  })

  it('shows the server error when the export is not ready', async () => {
    server.use(
      http.get(`${API}/exports/:id/download`, () =>
        apiError(409, 'export_not_ready', 'Выгрузка ещё не готова'),
      ),
    )
    renderButton()
    await userEvent.click(screen.getByRole('button', { name: 'Скачать CSV' }))
    expect(await screen.findByText('Выгрузка ещё не готова')).toBeInTheDocument()
  })
})
