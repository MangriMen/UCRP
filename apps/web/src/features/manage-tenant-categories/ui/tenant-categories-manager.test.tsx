import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it } from 'vitest'
import { TenantCategoriesManager } from './tenant-categories-manager'

function renderManager(canEdit = true) {
  render(
    <QueryClientProvider
      client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
    >
      <TenantCategoriesManager canEdit={canEdit} />
    </QueryClientProvider>,
  )
}

describe('TenantCategoriesManager', () => {
  it('hides editing controls for read-only roles', async () => {
    renderManager(false)
    expect(await screen.findByText('Питание')).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Добавить категорию' })).not.toBeInTheDocument()
  })

  it('validates required fields', async () => {
    renderManager()
    await userEvent.click(await screen.findByRole('button', { name: 'Добавить категорию' }))
    await userEvent.click(screen.getByRole('button', { name: 'Создать' }))
    expect(await screen.findByText('Введите код')).toBeInTheDocument()
  })

  it('shows a duplicate code error on the field', async () => {
    renderManager()
    await userEvent.click(await screen.findByRole('button', { name: 'Добавить категорию' }))
    const dialog = screen.getByRole('dialog')
    await userEvent.type(within(dialog).getByLabelText('Код'), 'food')
    await userEvent.type(within(dialog).getByLabelText('Название'), 'Еда')
    await userEvent.click(within(dialog).getByRole('button', { name: 'Создать' }))
    expect(
      await within(dialog).findByText('Категория с таким кодом уже существует'),
    ).toBeInTheDocument()
  })

  it('creates a category and shows it in the table', async () => {
    renderManager()
    await userEvent.click(await screen.findByRole('button', { name: 'Добавить категорию' }))
    const dialog = screen.getByRole('dialog')
    await userEvent.type(within(dialog).getByLabelText('Код'), 'sport')
    await userEvent.type(within(dialog).getByLabelText('Название'), 'Спорт')
    await userEvent.click(within(dialog).getByRole('button', { name: 'Создать' }))
    expect(await screen.findByText('Спорт')).toBeInTheDocument()
  })

  it('reports a category that is in use on delete', async () => {
    renderManager()
    await screen.findByText('Питание')
    await userEvent.click(screen.getAllByRole('button', { name: 'Удалить' })[0])
    await userEvent.click(
      within(screen.getByRole('alertdialog')).getByRole('button', { name: 'Удалить' }),
    )
    expect(await screen.findByText('Категория используется арендаторами')).toBeInTheDocument()
  })
})
