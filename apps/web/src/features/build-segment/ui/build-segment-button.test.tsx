import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it } from 'vitest'
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
})
