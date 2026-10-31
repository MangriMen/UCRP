import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, useLocation } from 'react-router'
import { describe, expect, it } from 'vitest'
import { useUrlTab } from './use-url-tab'

const TABS = ['overview', 'history'] as const

function Probe() {
  const [tab, setTab] = useUrlTab(TABS, 'overview')
  const location = useLocation()
  return (
    <div>
      <span data-testid="tab">{tab}</span>
      <span data-testid="search">{location.search}</span>
      <button onClick={() => setTab('history')}>history</button>
      <button onClick={() => setTab('overview')}>overview</button>
    </div>
  )
}

describe('useUrlTab', () => {
  it('falls back to the default tab for unknown values', () => {
    render(
      <MemoryRouter initialEntries={['/?tab=nope']}>
        <Probe />
      </MemoryRouter>,
    )
    expect(screen.getByTestId('tab')).toHaveTextContent('overview')
  })

  it('writes a non-default tab to the url and drops the default one', async () => {
    render(
      <MemoryRouter initialEntries={['/?q=a']}>
        <Probe />
      </MemoryRouter>,
    )
    await userEvent.click(screen.getByText('history'))
    expect(screen.getByTestId('tab')).toHaveTextContent('history')
    expect(screen.getByTestId('search')).toHaveTextContent('?q=a&tab=history')
    await userEvent.click(screen.getByText('overview'))
    expect(screen.getByTestId('search')).toHaveTextContent('?q=a')
  })
})
