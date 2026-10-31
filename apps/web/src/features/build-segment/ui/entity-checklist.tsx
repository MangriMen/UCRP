import { CheckboxField } from '@/shared/ui/checkbox-field'
import type { Option } from '@/shared/ui/option-select'

interface EntityChecklistProps {
  items: Option[]
  value: string[]
  onChange: (value: string[]) => void
}

export function EntityChecklist({ items, value, onChange }: EntityChecklistProps) {
  const toggle = (id: string, checked: boolean) =>
    onChange(checked ? [...value, id] : value.filter((current) => current !== id))

  if (items.length === 0) return <p className="text-xs text-muted-foreground">Список пуст</p>

  return (
    <div className="flex max-h-36 flex-col gap-2 overflow-y-auto rounded-md border p-2">
      {items.map((item) => (
        <CheckboxField
          key={item.value}
          label={item.label}
          checked={value.includes(item.value)}
          onCheckedChange={(checked) => toggle(item.value, checked)}
        />
      ))}
    </div>
  )
}
