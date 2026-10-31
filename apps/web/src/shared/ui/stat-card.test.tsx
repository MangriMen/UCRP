import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { StatCard } from './stat-card'

describe('StatCard', () => {
  it('shows label, value and hint', () => {
    render(<StatCard label="Визиты" value={42} hint="за 30 дней" />)
    expect(screen.getByText('Визиты')).toBeInTheDocument()
    expect(screen.getByText('42')).toBeInTheDocument()
    expect(screen.getByText('за 30 дней')).toBeInTheDocument()
  })

  it('hides the value while loading', () => {
    render(<StatCard label="Визиты" value={42} isLoading />)
    expect(screen.queryByText('42')).not.toBeInTheDocument()
  })
})
