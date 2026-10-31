import { useId } from 'react'
import { Checkbox } from './checkbox'
import { Label } from './label'

interface CheckboxFieldProps {
  label: string
  checked: boolean
  disabled?: boolean
  onCheckedChange: (checked: boolean) => void
}

export function CheckboxField({ label, checked, disabled, onCheckedChange }: CheckboxFieldProps) {
  const id = useId()
  return (
    <div className="flex items-center gap-2">
      <Checkbox id={id} checked={checked} disabled={disabled} onCheckedChange={onCheckedChange} />
      <Label htmlFor={id}>{label}</Label>
    </div>
  )
}
