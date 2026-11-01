import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import { describe, expect, it } from 'vitest'
import { ButtonLink } from './button'

describe('ButtonLink', () => {
  it('renders a real link styled as a button', () => {
    render(
      <MemoryRouter>
        <ButtonLink to="/tenants" variant="outline" className="w-full">
          Арендаторы
        </ButtonLink>
      </MemoryRouter>,
    )
    const link = screen.getByRole('link', { name: 'Арендаторы' })
    expect(link.tagName).toBe('A')
    expect(link).toHaveAttribute('href', '/tenants')
    expect(link).toHaveAttribute('data-slot', 'button')
    expect(link).toHaveClass('inline-flex', 'border-border', 'w-full')
    expect(screen.queryByRole('button')).not.toBeInTheDocument()
  })
})
