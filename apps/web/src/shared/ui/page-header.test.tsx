import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import { describe, expect, it } from 'vitest'
import { PageHeader } from './page-header'

describe('PageHeader', () => {
  it('renders breadcrumbs with links for parents and plain text for the current page', () => {
    render(
      <MemoryRouter>
        <PageHeader
          title="Анна"
          breadcrumbs={[{ label: 'Гости', to: '/guests' }, { label: 'Анна' }]}
        />
      </MemoryRouter>,
    )
    expect(screen.getByRole('link', { name: 'Гости' })).toHaveAttribute('href', '/guests')
    expect(screen.getByText('Анна', { selector: '[aria-current="page"]' })).toBeInTheDocument()
  })
})
