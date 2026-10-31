import {
  effectLabels,
  type AccessGrantItem,
  type AccessPermission,
  type GrantEffect,
  type GrantableRole,
} from '@/entities/access'
import { CheckboxField } from '@/shared/ui/checkbox-field'
import { OptionSelect } from '@/shared/ui/option-select'

const kindTitles: Record<AccessPermission['kind'], string> = {
  action: 'Действия',
  metric: 'Показатели',
  field: 'Поля ответа',
}

const KINDS: AccessPermission['kind'][] = ['action', 'metric', 'field']

const effectItems = (['allow', 'deny'] as const).map((effect) => ({
  value: effect,
  label: effectLabels[effect],
}))

interface PermissionChecklistProps {
  permissions: AccessPermission[]
  role: GrantableRole
  items: AccessGrantItem[]
  onChange: (items: AccessGrantItem[]) => void
}

export function PermissionChecklist({
  permissions,
  role,
  items,
  onChange,
}: PermissionChecklistProps) {
  const effectOf = (code: string) => items.find((item) => item.permission === code)?.effect
  const visible = permissions.filter(
    (permission) => permission.grantableTo.includes(role) || effectOf(permission.code),
  )

  const toggle = (code: string, checked: boolean) =>
    onChange(
      checked
        ? [...items, { permission: code, effect: 'allow' }]
        : items.filter((item) => item.permission !== code),
    )

  const setEffect = (code: string, effect: GrantEffect) =>
    onChange(items.map((item) => (item.permission === code ? { ...item, effect } : item)))

  if (visible.length === 0) {
    return (
      <p className="text-sm text-muted-foreground">
        У вас нет прав, которые можно выдать этому субъекту.
      </p>
    )
  }

  return (
    <div className="flex flex-col gap-4">
      {KINDS.map((kind) => {
        const group = visible.filter((permission) => permission.kind === kind)
        if (group.length === 0) return null
        return (
          <fieldset key={kind} className="flex flex-col gap-2">
            <legend className="mb-1 text-sm font-medium">{kindTitles[kind]}</legend>
            {group.map((permission) => {
              const effect = effectOf(permission.code)
              const grantable = permission.grantableTo.includes(role)
              return (
                <div key={permission.code} className="flex flex-wrap items-center gap-x-3 gap-y-1">
                  <CheckboxField
                    label={permission.title}
                    checked={effect !== undefined}
                    onCheckedChange={(checked) => toggle(permission.code, checked)}
                  />
                  <code className="text-xs text-muted-foreground">{permission.code}</code>
                  {effect && (
                    <OptionSelect
                      ariaLabel={`Эффект: ${permission.title}`}
                      className="h-8 w-36"
                      items={effectItems}
                      value={effect}
                      onValueChange={(next) => setEffect(permission.code, next as GrantEffect)}
                    />
                  )}
                  {!grantable && (
                    <span className="text-xs text-destructive">Нельзя выдать от вашего имени</span>
                  )}
                </div>
              )
            })}
          </fieldset>
        )
      })}
    </div>
  )
}
