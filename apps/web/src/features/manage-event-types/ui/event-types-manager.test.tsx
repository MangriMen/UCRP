import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it } from 'vitest'
import { EventTypesManager } from './event-types-manager'

function renderManager() {
  render(
    <QueryClientProvider
      client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
    >
      <EventTypesManager canEdit />
    </QueryClientProvider>,
  )
}

describe('EventTypesManager', () => {
  it('rejects a duplicate code on the field', async () => {
    renderManager()
    await userEvent.click(await screen.findByRole('button', { name: 'Добавить тип события' }))
    const dialog = screen.getByRole('dialog')
    await userEvent.type(within(dialog).getByLabelText('Код'), 'purchase')
    await userEvent.type(within(dialog).getByLabelText('Название'), 'Ещё покупка')
    await userEvent.click(within(dialog).getByRole('button', { name: 'Создать' }))
    expect(
      await within(dialog).findByText('Тип события с таким кодом уже существует'),
    ).toBeInTheDocument()
  })

  it('refuses to delete an event type in use', async () => {
    renderManager()
    await screen.findByText('Покупка')
    await userEvent.click(screen.getAllByRole('button', { name: 'Удалить' })[0])
    await userEvent.click(
      within(screen.getByRole('alertdialog')).getByRole('button', { name: 'Удалить' }),
    )
    expect(await screen.findByText('Тип события используется в событиях')).toBeInTheDocument()
  })
})
