import { OptionSelect, type Option } from '@/shared/ui/option-select'
import type { SubjectDirectory } from '../api/use-subject-directory'
import { subjectKindLabels, type SubjectKind, type SubjectValue } from '../model/subject'

const KINDS: SubjectKind[] = ['tenant', 'tenant_user', 'tenant_role', 'manager']
const NOT_SELECTED = '__none__'

interface SubjectPickerProps {
  value: SubjectValue
  onChange: (value: SubjectValue) => void
  directory: SubjectDirectory
  excludeSelf?: boolean
  invalid?: boolean
}

function subjectOptions(
  kind: SubjectKind,
  directory: SubjectDirectory,
  excludeSelf: boolean,
): Option[] {
  if (kind === 'tenant') return directory.tenants.map(({ id, label }) => ({ value: id, label }))
  const role = kind === 'manager' ? 'manager' : 'tenant'
  return directory.users
    .filter((user) => user.role === role && !(excludeSelf && user.id === directory.selfId))
    .map(({ id, label }) => ({ value: id, label }))
}

export function SubjectPicker({
  value,
  onChange,
  directory,
  excludeSelf = false,
  invalid,
}: SubjectPickerProps) {
  const kindItems = KINDS.map((kind) => ({ value: kind, label: subjectKindLabels[kind] }))
  const options = subjectOptions(value.kind, directory, excludeSelf)
  const items = [
    { value: NOT_SELECTED, label: options.length ? 'Выберите…' : 'Нет вариантов' },
    ...options,
  ]

  return (
    <div className="flex flex-wrap gap-3">
      <OptionSelect
        ariaLabel="Тип субъекта"
        className="w-56"
        items={kindItems}
        value={value.kind}
        onValueChange={(kind) => onChange({ kind: kind as SubjectKind, id: '' })}
      />
      {value.kind !== 'tenant_role' && (
        <OptionSelect
          ariaLabel="Субъект"
          className="w-64"
          invalid={invalid}
          items={items}
          value={value.id || NOT_SELECTED}
          onValueChange={(id) => onChange({ ...value, id: id === NOT_SELECTED ? '' : id })}
        />
      )}
    </div>
  )
}
