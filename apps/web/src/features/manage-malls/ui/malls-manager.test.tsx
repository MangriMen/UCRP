import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it } from 'vitest'
import { MallsManager } from './malls-manager'

function renderManager(canEdit = true) {
  render(
    <QueryClientProvider
      client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
    >
      <MallsManager canEdit={canEdit} />
    </QueryClientProvider>,
  )
}

describe('MallsManager', () => {
  it('lists malls with floors and hides editing controls for read-only roles', async () => {
    renderManager(false)
    expect(await screen.findByText('Центральный ТЦ')).toBeInTheDocument()
    expect(screen.getByText('-1, 1, 2, 3, 4')).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Добавить ТЦ' })).not.toBeInTheDocument()
  })

  it('validates required fields', async () => {
    renderManager()
    await userEvent.click(await screen.findByRole('button', { name: 'Добавить ТЦ' }))
    await userEvent.click(screen.getByRole('button', { name: 'Создать' }))
    expect(await screen.findByText('Введите код')).toBeInTheDocument()
  })

  it('shows a duplicate code error on the field', async () => {
    renderManager()
    await userEvent.click(await screen.findByRole('button', { name: 'Добавить ТЦ' }))
    const dialog = screen.getByRole('dialog')
    await userEvent.type(within(dialog).getByLabelText('Код'), 'central')
    await userEvent.type(within(dialog).getByLabelText('Название'), 'Другой ТЦ')
    await userEvent.click(within(dialog).getByRole('button', { name: 'Создать' }))
    expect(await within(dialog).findByText('ТЦ с таким кодом уже существует')).toBeInTheDocument()
  })

  it('creates a mall and shows it in the table', async () => {
    renderManager()
    await userEvent.click(await screen.findByRole('button', { name: 'Добавить ТЦ' }))
    const dialog = screen.getByRole('dialog')
    await userEvent.type(within(dialog).getByLabelText('Код'), 'south')
    await userEvent.type(within(dialog).getByLabelText('Название'), 'Южный ТЦ')
    await userEvent.click(within(dialog).getByRole('button', { name: 'Создать' }))
    expect(await screen.findByText('Южный ТЦ')).toBeInTheDocument()
  })

  it('reports a mall with tenants on delete', async () => {
    renderManager()
    await screen.findByText('Центральный ТЦ')
    await userEvent.click(screen.getAllByRole('button', { name: 'Удалить' })[0])
    await userEvent.click(
      within(screen.getByRole('alertdialog')).getByRole('button', { name: 'Удалить' }),
    )
    expect(await screen.findByText('У ТЦ есть арендаторы, удаление невозможно')).toBeInTheDocument()
  })
})
