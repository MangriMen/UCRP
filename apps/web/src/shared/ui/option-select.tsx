import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from './select'

export interface Option {
  value: string
  label: string
}

interface OptionSelectProps {
  items: Option[]
  value: string
  onValueChange: (value: string) => void
  id?: string
  ariaLabel?: string
  invalid?: boolean
  className?: string
}

export function OptionSelect({
  items,
  value,
  onValueChange,
  id,
  ariaLabel,
  invalid,
  className,
}: OptionSelectProps) {
  return (
    <Select items={items} value={value} onValueChange={(next) => next && onValueChange(next)}>
      <SelectTrigger id={id} aria-label={ariaLabel} aria-invalid={invalid} className={className}>
        <SelectValue />
      </SelectTrigger>
      <SelectContent>
        {items.map((item) => (
          <SelectItem key={item.value} value={item.value}>
            {item.label}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  )
}
