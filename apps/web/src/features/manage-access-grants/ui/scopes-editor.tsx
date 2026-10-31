import { XIcon } from 'lucide-react'
import { scopeTypeLabels } from '@/entities/access'
import { Button } from '@/shared/ui/button'
import { Input } from '@/shared/ui/input'
import { OptionSelect } from '@/shared/ui/option-select'
import { EXTRA_SCOPE_TYPES, type ExtraScopeType, type ScopeDraft } from '../model/grant-form'

const typeItems = EXTRA_SCOPE_TYPES.map((type) => ({ value: type, label: scopeTypeLabels[type] }))

const placeholders: Record<ExtraScopeType, string> = {
  category: 'ID категории',
  period_days: 'Например, 90',
  channel: 'sms, push или email',
  k_threshold: 'Не меньше 5',
  row_limit: 'Например, 1000',
  field_mask: 'Код поля',
  metric: 'Код показателя',
}

interface ScopesEditorProps {
  scopes: ScopeDraft[]
  onChange: (scopes: ScopeDraft[]) => void
}

export function ScopesEditor({ scopes, onChange }: ScopesEditorProps) {
  const change = (index: number, patch: Partial<ScopeDraft>) =>
    onChange(scopes.map((scope, current) => (current === index ? { ...scope, ...patch } : scope)))

  return (
    <div className="flex flex-col gap-2">
      {scopes.map((scope, index) => (
        <div key={index} className="flex flex-wrap items-center gap-2">
          <OptionSelect
            ariaLabel={`Тип ограничения ${index + 1}`}
            className="w-56"
            items={typeItems}
            value={scope.type}
            onValueChange={(type) => change(index, { type: type as ExtraScopeType })}
          />
          <Input
            aria-label={`Значение ограничения ${index + 1}`}
            className="w-56"
            placeholder={placeholders[scope.type]}
            value={scope.value}
            onChange={(event) => change(index, { value: event.target.value })}
          />
          <Button
            type="button"
            variant="ghost"
            size="icon"
            aria-label={`Убрать ограничение ${index + 1}`}
            onClick={() => onChange(scopes.filter((_, current) => current !== index))}
          >
            <XIcon />
          </Button>
        </div>
      ))}
      <div>
        <Button
          type="button"
          variant="outline"
          size="sm"
          onClick={() => onChange([...scopes, { type: 'period_days', value: '' }])}
        >
          Добавить ограничение
        </Button>
      </div>
    </div>
  )
}
