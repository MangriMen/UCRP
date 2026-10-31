import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useState } from 'react'
import { describe, expect, it } from 'vitest'
import { TenantCategoriesManager } from './tenant-categories-manager'

function ManagerWithRelationsState({ canEdit }: { canEdit: boolean }) {
  const [relationsCategoryId, setRelationsCategoryId] = useState<string | null>(null)
  return (
    <TenantCategoriesManager
      canEdit={canEdit}
      relationsCategoryId={relationsCategoryId}
      onRelationsCategoryChange={setRelationsCategoryId}
    />
  )
}

function renderManager(canEdit = true) {
  render(
    <QueryClientProvider
      client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
    >
      <ManagerWithRelationsState canEdit={canEdit} />
    </QueryClientProvider>,
  )
}

const rowOf = (text: string) => within(screen.getByText(text).closest('tr') as HTMLElement)

async function openRelations(categoryName: string) {
  await screen.findByText(categoryName)
  await userEvent.click(rowOf(categoryName).getByRole('button', { name: 'Связи' }))
  return within(await screen.findByRole('dialog'))
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

  it('shows relations in both directions read-only for roles without write access', async () => {
    renderManager(false)
    const dialog = await openRelations('Питание')
    expect(await dialog.findByText('Развлечения')).toBeInTheDocument()
    expect(dialog.getByText('Смежная')).toBeInTheDocument()
    expect(dialog.getByText('70')).toBeInTheDocument()
    expect(dialog.queryByRole('button', { name: 'Сохранить' })).not.toBeInTheDocument()
    expect(dialog.queryByRole('button', { name: 'Добавить связь' })).not.toBeInTheDocument()
  })

  it('shows an empty state when the category has no relations', async () => {
    renderManager(false)
    const dialog = await openRelations('Одежда и обувь')
    expect(await dialog.findByText('Связей пока нет')).toBeInTheDocument()
  })

  it('replaces the relation set and shows it from the related side', async () => {
    renderManager()
    let dialog = await openRelations('Развлечения')
    expect(await dialog.findByLabelText('Вес в процентах, связь 1')).toHaveValue(70)
    await userEvent.click(dialog.getByRole('button', { name: 'Удалить связь с «Питание»' }))
    expect(dialog.getByText('Связей пока нет')).toBeInTheDocument()
    await userEvent.click(dialog.getByRole('button', { name: 'Добавить связь' }))
    const weight = dialog.getByLabelText('Вес в процентах, связь 1')
    await userEvent.clear(weight)
    await userEvent.type(weight, '120')
    expect(dialog.getByText('Вес от 0 до 100')).toBeInTheDocument()
    expect(dialog.getByRole('button', { name: 'Сохранить' })).toBeDisabled()
    await userEvent.clear(weight)
    await userEvent.type(weight, '55')
    await userEvent.click(dialog.getByRole('button', { name: 'Сохранить' }))
    await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument())

    dialog = await openRelations('Питание')
    expect(await dialog.findByDisplayValue('55')).toBeInTheDocument()
  })
})
