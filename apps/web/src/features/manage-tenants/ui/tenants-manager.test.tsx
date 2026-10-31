import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it } from 'vitest'
import { TenantsManager } from './tenants-manager'

function renderManager(canEdit = true) {
  render(
    <QueryClientProvider
      client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
    >
      <TenantsManager canEdit={canEdit} />
    </QueryClientProvider>,
  )
}

describe('TenantsManager', () => {
  it('lists tenants with their category', async () => {
    renderManager(false)
    expect(await screen.findByText('Бистро «Уголок»')).toBeInTheDocument()
    expect(screen.getByRole('columnheader', { name: 'ТЦ' })).toBeInTheDocument()
    expect((await screen.findAllByText('Центральный ТЦ')).length).toBeGreaterThan(0)
    expect(screen.queryByRole('button', { name: 'Добавить арендатора' })).not.toBeInTheDocument()
  })

  it('validates the form and creates a tenant', async () => {
    renderManager()
    await userEvent.click(await screen.findByRole('button', { name: 'Добавить арендатора' }))
    const dialog = screen.getByRole('dialog')
    await userEvent.click(within(dialog).getByRole('button', { name: 'Создать' }))
    expect(await within(dialog).findByText('Введите код')).toBeInTheDocument()
    expect(within(dialog).getByText('Выберите категорию')).toBeInTheDocument()
    expect(within(dialog).getByText('Выберите ТЦ')).toBeInTheDocument()

    await userEvent.type(within(dialog).getByLabelText('Код'), 'books')
    await userEvent.type(within(dialog).getByLabelText('Название'), 'Книжный')
    await userEvent.type(within(dialog).getByLabelText('Этаж'), '2')
    await userEvent.click(within(dialog).getByLabelText('Категория'))
    await userEvent.click(await screen.findByRole('option', { name: 'Питание' }))
    await userEvent.click(within(dialog).getByLabelText('ТЦ'))
    await userEvent.click(await screen.findByRole('option', { name: 'Северный ТЦ' }))
    await userEvent.click(within(dialog).getByRole('button', { name: 'Создать' }))
    expect(await screen.findByText('Книжный')).toBeInTheDocument()
  })

  it('shows the conflict when deleting a tenant with events', async () => {
    renderManager()
    await screen.findByText('Заря')
    await userEvent.click(screen.getAllByRole('button', { name: 'Удалить' })[0])
    await userEvent.click(
      within(screen.getByRole('alertdialog')).getByRole('button', { name: 'Удалить' }),
    )
    expect(
      await screen.findByText('У арендатора есть события, удаление невозможно'),
    ).toBeInTheDocument()
  })
})
