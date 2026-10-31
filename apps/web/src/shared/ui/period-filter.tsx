import type { ReactNode } from 'react'
import type { Period } from '@/shared/lib/period'
import { FormField } from './form-field'
import { Input } from './input'

interface PeriodFilterProps {
  value: Period
  onChange: (period: Period) => void
  error?: string
  children?: ReactNode
}

export function PeriodFilter({ value, onChange, error, children }: PeriodFilterProps) {
  return (
    <div className="flex flex-wrap items-start gap-3">
      <FormField label="С" error={error}>
        {(props) => (
          <Input
            {...props}
            type="date"
            className="w-44"
            value={value.from}
            onChange={(event) => onChange({ ...value, from: event.target.value })}
          />
        )}
      </FormField>
      <FormField label="По">
        {(props) => (
          <Input
            {...props}
            type="date"
            className="w-44"
            value={value.to}
            onChange={(event) => onChange({ ...value, to: event.target.value })}
          />
        )}
      </FormField>
      {children}
    </div>
  )
}
