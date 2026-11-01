import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes, useLocation } from 'react-router'
import { afterEach, beforeEach, describe, expect, it } from 'vitest'
import { mockUsers } from '@/entities/user/mocks'
import { setAccessToken } from '@/shared/api/client'
import { TenantsPage } from './tenants-page'

const SESSION_KEY = 'ucrp-mock-session'
const [marketer, , , analyst] = mockUsers

function Search() {
  return <span data-testid="search">{useLocation().search}</span>
}

function renderPage(url = '/tenants') {
  render(
    <QueryClientProvider
      client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
    >
      <MemoryRouter initialEntries={[url]}>
        <Routes>
          <Route
            path="/tenants"
            element={
              <>
                <TenantsPage />
                <Search />
              </>
            }
          />
          <Route path="/tenants/:id" element={<div>Карточка арендатора</div>} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

const rowTexts = () =>
  screen
    .getAllByRole('row')
    .slice(1)
    .map((row) => row.textContent ?? '')

describe('TenantsPage', () => {
  beforeEach(() => {
    localStorage.setItem(SESSION_KEY, marketer.userId)
  })

  afterEach(() => {
    localStorage.clear()
    setAccessToken(null)
  })

  it('lists tenants with statistics and marks suppressed values and problem sources', async () => {
    renderPage()
    expect(await screen.findByText('Бистро «Уголок»')).toBeInTheDocument()
    const bistro = within(screen.getByText('Бистро «Уголок»').closest('tr') as HTMLElement)
    expect(bistro.getByText('48')).toBeInTheDocument()
    expect(bistro.getByText('130')).toBeInTheDocument()
    const zarya = within(screen.getByText('Заря').closest('tr') as HTMLElement)
    expect(zarya.getAllByText('меньше 5')).toHaveLength(3)
    expect(zarya.getByText('Проблемы: 1')).toBeInTheDocument()
  })

  it('writes the sort of a numeric column to the URL and orders rows', async () => {
    renderPage()
    const sortButton = await screen.findByRole('button', { name: /Визитов/ })
    await userEvent.click(sortButton)
    await userEvent.click(sortButton)
    expect(screen.getByTestId('search')).toHaveTextContent('sort=-visitsCount')
    await screen.findByText('Киномир')
    expect(rowTexts()[0]).toContain('Киномир')
    expect(rowTexts()[2]).toContain('Заря')
  })

  it('restores filters from the URL and resets them', async () => {
    renderPage('/tenants?is_active=false')
    expect(await screen.findByText('Киномир')).toBeInTheDocument()
    expect(screen.queryByText('Заря')).not.toBeInTheDocument()
    await userEvent.click(screen.getByRole('button', { name: 'Сбросить' }))
    expect(await screen.findByText('Заря')).toBeInTheDocument()
  })

  it('filters by mall and keeps the mall in the URL', async () => {
    renderPage()
    await screen.findByText('Заря')
    await userEvent.click(screen.getByLabelText('ТЦ'))
    await userEvent.click(await screen.findByRole('option', { name: 'Северный ТЦ' }))
    expect(screen.getByTestId('search')).toHaveTextContent('mall_id=mall-0002')
    expect(await screen.findByText('Арендаторы по фильтрам не найдены')).toBeInTheDocument()
  })

  it('restores the mall filter from the URL', async () => {
    renderPage('/tenants?mall_id=mall-0001')
    expect(await screen.findByText('Заря')).toBeInTheDocument()
    expect(screen.getByLabelText('ТЦ')).toHaveTextContent('Центральный ТЦ')
  })

  it('searches by name', async () => {
    renderPage()
    await userEvent.type(screen.getByLabelText('Поиск арендаторов'), 'кино')
    expect(await screen.findByText('Киномир')).toBeInTheDocument()
    await waitFor(() => expect(screen.queryByText('Заря')).not.toBeInTheDocument())
  })

  it('opens the tenant card on row click', async () => {
    renderPage()
    await userEvent.click(await screen.findByText('Заря'))
    expect(await screen.findByText('Карточка арендатора')).toBeInTheDocument()
  })

  it('hides tenant creation without reference.write', async () => {
    renderPage()
    await screen.findByText('Заря')
    expect(screen.queryByRole('button', { name: 'Добавить арендатора' })).not.toBeInTheDocument()
  })
})

describe('TenantsPage tenant creation', () => {
  beforeEach(() => {
    localStorage.setItem(SESSION_KEY, analyst.userId)
  })

  afterEach(() => {
    localStorage.clear()
    setAccessToken(null)
  })

  it('validates the form and creates a tenant', async () => {
    renderPage()
    await userEvent.click(await screen.findByRole('button', { name: 'Добавить арендатора' }))
    const dialog = screen.getByRole('dialog')
    await userEvent.click(within(dialog).getByRole('button', { name: 'Создать' }))
    expect(await within(dialog).findByText('Введите код')).toBeInTheDocument()
    expect(within(dialog).getByText('Выберите категорию')).toBeInTheDocument()

    await userEvent.type(within(dialog).getByLabelText('Код'), 'books')
    await userEvent.type(within(dialog).getByLabelText('Название'), 'Книжный')
    await userEvent.type(within(dialog).getByLabelText('Этаж'), '2')
    await userEvent.click(within(dialog).getByLabelText('Категория'))
    await userEvent.click(await screen.findByRole('option', { name: 'Питание' }))
    await userEvent.click(within(dialog).getByLabelText('ТЦ'))
    await userEvent.click((await screen.findAllByRole('option'))[0])
    await userEvent.click(within(dialog).getByRole('button', { name: 'Создать' }))
    expect(await screen.findByText('Книжный')).toBeInTheDocument()
  })

  it('drops an unknown mall from the URL once malls are loaded', async () => {
    renderPage('/tenants?mall_id=xxx&q=Заря')
    await waitFor(() => expect(screen.getByTestId('search')).not.toHaveTextContent('mall_id'))
    expect(screen.getByTestId('search')).toHaveTextContent('q=')
    expect(screen.getByRole('combobox', { name: 'ТЦ' })).toHaveTextContent('Все ТЦ')
  })

  it('reports a duplicate code on the code field', async () => {
    renderPage()
    await userEvent.click(await screen.findByRole('button', { name: 'Добавить арендатора' }))
    const dialog = screen.getByRole('dialog')
    await userEvent.type(within(dialog).getByLabelText('Код'), 'ZARA')
    await userEvent.type(within(dialog).getByLabelText('Название'), 'Дубль')
    await userEvent.click(within(dialog).getByLabelText('Категория'))
    await userEvent.click(await screen.findByRole('option', { name: 'Питание' }))
    await userEvent.click(within(dialog).getByLabelText('ТЦ'))
    await userEvent.click((await screen.findAllByRole('option'))[0])
    await userEvent.click(within(dialog).getByRole('button', { name: 'Создать' }))
    expect(
      await within(dialog).findByText('Арендатор с таким кодом уже существует'),
    ).toBeInTheDocument()
  })
})
