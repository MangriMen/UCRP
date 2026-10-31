import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { http, HttpResponse } from 'msw'
import { describe, expect, it, vi } from 'vitest'
import { server } from '@/app/mocks/server'
import { API } from '@/shared/api/mocks/helpers'
import { UploadImportForm } from './upload-import-form'

function renderForm(onUploaded = vi.fn()) {
  render(
    <QueryClientProvider client={new QueryClient()}>
      <UploadImportForm onUploaded={onUploaded} />
    </QueryClientProvider>,
  )
  return onUploaded
}

const chooseFile = (file: File) =>
  userEvent.upload(screen.getByLabelText('Файл для импорта'), file, { applyAccept: false })

describe('UploadImportForm', () => {
  it('rejects an unsupported file and keeps submit disabled', async () => {
    renderForm()
    await chooseFile(new File(['x'], 'report.pdf'))
    expect(await screen.findByText(/Поддерживаются файлы CSV и Excel/)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Загрузить' })).toBeDisabled()
  })

  it('uploads the chosen file for the selected source', async () => {
    server.use(
      http.post(`${API}/imports`, () =>
        HttpResponse.json(
          { batchId: 'batch-new', status: 'queued', duplicate: false },
          { status: 202 },
        ),
      ),
    )
    const onUploaded = renderForm()
    await userEvent.click(screen.getByRole('combobox'))
    await userEvent.click(await screen.findByRole('option', { name: 'CRM «Лента»' }))
    await chooseFile(new File(['id,date'], 'fresh-file.csv'))
    await userEvent.click(screen.getByRole('button', { name: 'Загрузить' }))
    await vi.waitFor(() => expect(onUploaded).toHaveBeenCalledWith('batch-new', false))
  })

  it('reports an already uploaded file as a duplicate', async () => {
    server.use(
      http.post(`${API}/imports`, () =>
        HttpResponse.json({ batchId: 'batch-0001', status: 'completed', duplicate: true }),
      ),
    )
    const onUploaded = renderForm()
    await userEvent.click(screen.getByRole('combobox'))
    await userEvent.click(await screen.findByRole('option', { name: 'CRM «Лента»' }))
    await chooseFile(new File(['id,date'], 'crm-october.csv'))
    await userEvent.click(screen.getByRole('button', { name: 'Загрузить' }))
    await vi.waitFor(() => expect(onUploaded).toHaveBeenCalledWith('batch-0001', true))
  })
})
