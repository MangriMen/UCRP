import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes, useParams } from 'react-router'
import { describe, expect, it } from 'vitest'
import { userKeys } from '@/entities/user'
import { mockUsers } from '@/entities/user/mocks'
import { AccessGrantCreatePage, AccessGrantEditPage } from './access-grant-form-page'

const manager = mockUsers[1]

function Card() {
  return <div>{`Карточка ${useParams().id}`}</div>
}

function renderPage(url: string) {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  client.setQueryData(userKeys.me, manager)
  render(
    <QueryClientProvider client={client}>
      <MemoryRouter initialEntries={[url]}>
        <Routes>
          <Route path="/access/grants/new" element={<AccessGrantCreatePage />} />
          <Route path="/access/grants/:id/edit" element={<AccessGrantEditPage />} />
          <Route path="/access/grants/:id" element={<Card />} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

const choose = async (label: string, option: string) => {
  await userEvent.click(await screen.findByLabelText(label))
  await userEvent.click(await screen.findByRole('option', { name: option }))
}

describe('AccessGrantCreatePage', () => {
  it('validates the form before sending it', async () => {
    renderPage('/access/grants/new')
    await userEvent.click(await screen.findByRole('button', { name: 'Выдать доступ' }))
    expect(await screen.findByText('Выберите субъект')).toBeInTheDocument()
    expect(screen.getByText('Выберите хотя бы одно право')).toBeInTheDocument()
    expect(screen.getByText('Укажите причину')).toBeInTheDocument()
  })

  it('offers only permissions the issuer can delegate to a tenant', async () => {
    renderPage('/access/grants/new')
    expect(await screen.findByText('Число визитов')).toBeInTheDocument()
    expect(screen.queryByText('Суммарная выручка')).not.toBeInTheDocument()
    expect(screen.queryByText('Поиск гостей')).not.toBeInTheDocument()
    expect(screen.queryByText('Выдача и отзыв грантов')).not.toBeInTheDocument()
  })

  it('applies a preset, simulates the result and creates the grant for the subject from the URL', async () => {
    renderPage('/access/grants/new?subject_type=tenant&subject_id=tenant-0003')
    await choose('Пресет', 'Свои агрегаты')
    await userEvent.type(screen.getByLabelText('Причина'), 'Пилот с кинотеатром')
    await userEvent.click(screen.getByRole('button', { name: 'Проверить доступ' }))
    const result = (await screen.findByText('Что получит субъект')).closest(
      '[data-slot="card"]',
    ) as HTMLElement
    expect(within(result).getByText('Число визитов')).toBeInTheDocument()
    await userEvent.click(screen.getByRole('button', { name: 'Выдать доступ' }))
    expect(await screen.findByText(/^Карточка grant-/)).toBeInTheDocument()
  })

  it('limits a manager to personal-data permissions without presets and scopes', async () => {
    renderPage('/access/grants/new')
    await choose('Тип субъекта', 'Управляющий')
    expect(await screen.findByText(/только права на личные данные гостей/)).toBeInTheDocument()
    expect(screen.queryByLabelText('Пресет')).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Добавить ограничение' })).not.toBeInTheDocument()
    expect(screen.getByText('Поиск гостей')).toBeInTheDocument()
    expect(screen.queryByText('Число визитов')).not.toBeInTheDocument()
  })
})

describe('AccessGrantEditPage', () => {
  it('saves a changed reason and returns to the card', async () => {
    renderPage('/access/grants/grant-0002/edit')
    const reason = await screen.findByLabelText('Причина')
    expect(reason).toHaveValue('События своих гостей')
    await userEvent.clear(reason)
    await userEvent.type(reason, 'События за квартал')
    await userEvent.click(screen.getByRole('button', { name: 'Сохранить' }))
    expect(await screen.findByText('Карточка grant-0002')).toBeInTheDocument()
  })

  it('refuses to edit a revoked grant', async () => {
    renderPage('/access/grants/grant-0006/edit')
    expect(await screen.findByText('Грант нельзя изменить')).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Сохранить' })).not.toBeInTheDocument()
  })
})
