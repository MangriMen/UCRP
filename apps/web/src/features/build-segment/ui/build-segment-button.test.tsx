import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { http, HttpResponse } from 'msw'
import { describe, expect, it } from 'vitest'
import { server } from '@/app/mocks/server'
import { API } from '@/shared/api/mocks/helpers'
import { mockSegments } from '@/entities/segment/mocks'
import { BuildSegmentButton } from './build-segment-button'

function renderButton(segment?: (typeof mockSegments)[number]) {
  render(
    <QueryClientProvider
      client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
    >
      <BuildSegmentButton segment={segment} />
    </QueryClientProvider>,
  )
}

describe('BuildSegmentButton', () => {
  it('validates the name and the period before creating', async () => {
    renderButton()
    await userEvent.click(screen.getByRole('button', { name: 'Новый сегмент' }))
    const dialog = await screen.findByRole('dialog')
    await userEvent.clear(within(dialog).getByLabelText('Период анализа, дней'))
    await userEvent.type(within(dialog).getByLabelText('Период анализа, дней'), '900')
    await userEvent.click(within(dialog).getByRole('button', { name: 'Создать сегмент' }))
    expect(await within(dialog).findByText('Введите название')).toBeInTheDocument()
    expect(within(dialog).getByText('Целое число от 1 до 730')).toBeInTheDocument()
  })

  it('closes the dialog after a new version is saved', async () => {
    renderButton(mockSegments[0])
    await userEvent.click(screen.getByRole('button', { name: 'Изменить правило' }))
    const dialog = await screen.findByRole('dialog')
    expect(within(dialog).getByLabelText('Название')).toHaveValue(mockSegments[0]?.name)
    await userEvent.click(within(dialog).getByRole('button', { name: 'Сохранить новую версию' }))
    await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument())
  })

  it('prefills the form from a template', async () => {
    renderButton()
    await userEvent.click(screen.getByRole('button', { name: 'Новый сегмент' }))
    const dialog = await screen.findByRole('dialog')
    await userEvent.click(await within(dialog).findByRole('button', { name: /Риск оттока/ }))
    expect(within(dialog).getByLabelText('Название')).toHaveValue('Риск оттока')
    expect(within(dialog).getByLabelText('Период анализа, дней')).toHaveValue('60')
    expect(within(dialog).getByLabelText('Минимум визитов')).toHaveValue('2')
  })

  it('sends the selected purchase zones to the audience preview and on save', async () => {
    let previewBody: unknown
    let createBody: unknown
    server.use(
      http.post(`${API}/segments/preview`, async ({ request }) => {
        previewBody = await request.json()
        return HttpResponse.json({ size: 42, examples: [], cached: false, suppressed: false })
      }),
      http.post(`${API}/segments`, async ({ request }) => {
        createBody = await request.json()
        return HttpResponse.json(mockSegments[0], { status: 201 })
      }),
    )
    renderButton()
    await userEvent.click(screen.getByRole('button', { name: 'Новый сегмент' }))
    const dialog = await screen.findByRole('dialog')
    await userEvent.type(within(dialog).getByLabelText('Название'), 'Покупатели')
    await userEvent.click(
      await within(dialog).findByRole('checkbox', { name: 'Центральный ТЦ · 2 этаж' }),
    )
    await userEvent.click(within(dialog).getByRole('checkbox', { name: 'Северный ТЦ · 2 этаж' }))

    const zones = [
      { mallId: 'mall-0001', floor: 2 },
      { mallId: 'mall-0002', floor: 2 },
    ]
    await waitFor(() =>
      expect(previewBody).toMatchObject({
        rule: { conditions: [{ periodDays: 90, purchasedZones: zones }] },
      }),
    )
    expect(await within(dialog).findByText('42')).toBeInTheDocument()

    await userEvent.click(within(dialog).getByRole('button', { name: 'Создать сегмент' }))
    await waitFor(() =>
      expect(createBody).toMatchObject({
        rule: { conditions: [{ periodDays: 90, purchasedZones: zones }] },
      }),
    )
  })

  it('submits a root group with a nested OR group', async () => {
    let body: unknown
    server.use(
      http.post(`${API}/segments`, async ({ request }) => {
        body = await request.json()
        return HttpResponse.json(mockSegments[0], { status: 201 })
      }),
    )
    renderButton()
    await userEvent.click(screen.getByRole('button', { name: 'Новый сегмент' }))
    const dialog = await screen.findByRole('dialog')
    await userEvent.type(within(dialog).getByLabelText('Название'), 'Группы')
    await userEvent.click(within(dialog).getByRole('button', { name: 'Добавить группу' }))
    expect(
      within(dialog.querySelector('form') as HTMLElement).getAllByText('Удалить группу'),
    ).toHaveLength(1)
    const orButtons = within(dialog).getAllByRole('button', { name: 'Любое условие (ИЛИ)' })
    expect(orButtons).toHaveLength(2)
    await userEvent.click(orButtons[1] as HTMLElement)
    expect(orButtons[1]).toHaveAttribute('aria-pressed', 'true')
    expect(orButtons[0]).toHaveAttribute('aria-pressed', 'false')
    await userEvent.click(within(dialog).getByRole('button', { name: 'Создать сегмент' }))
    await waitFor(() =>
      expect(body).toMatchObject({
        name: 'Группы',
        rule: {
          operator: 'and',
          conditions: [{ periodDays: 90 }, { operator: 'or', conditions: [{ periodDays: 90 }] }],
        },
      }),
    )
  })
})
