import { get, useFieldArray, useFormContext } from 'react-hook-form'
import { Button } from '@/shared/ui/button'
import {
  emptyGroupValues,
  emptyLeafValues,
  operatorOptions,
  type Operator,
  type SegmentFormValues,
} from '../model/schema'
import { LeafFields } from './leaf-fields'

const operatorLabels: Record<Operator, string> = {
  and: 'Все условия (И)',
  or: 'Любое условие (ИЛИ)',
}

interface ArrayError {
  message?: string
  root?: { message?: string }
}

interface GroupEditorProps {
  path: string
  remainingDepth: number
  onRemove?: () => void
}

const join = (path: string, name: string) => (path ? `${path}.${name}` : name)

export function GroupEditor({ path, remainingDepth, onRemove }: GroupEditorProps) {
  const { control, watch, setValue, formState } = useFormContext<SegmentFormValues>()
  const operatorName = join(path, 'operator') as 'operator'
  const operator = watch(operatorName)
  const conditions = useFieldArray({ control, name: join(path, 'conditions') as 'conditions' })
  const groups = useFieldArray({ control, name: join(path, 'groups') as 'groups' })
  const conditionsError: ArrayError | undefined = get(formState.errors, join(path, 'conditions'))
  const message = conditionsError?.message ?? conditionsError?.root?.message

  return (
    <div className="flex flex-col gap-4 rounded-md border p-3">
      <div className="flex flex-wrap items-center gap-2">
        {operatorOptions.map((option) => (
          <Button
            key={option}
            type="button"
            size="sm"
            variant={operator === option ? 'secondary' : 'outline'}
            aria-pressed={operator === option}
            onClick={() => setValue(operatorName, option, { shouldDirty: true })}
          >
            {operatorLabels[option]}
          </Button>
        ))}
        {onRemove && (
          <Button type="button" size="sm" variant="ghost" className="ml-auto" onClick={onRemove}>
            Удалить группу
          </Button>
        )}
      </div>
      {conditions.fields.map((field, index) => (
        <div key={field.id} className="flex flex-col gap-3 rounded-md bg-muted/30 p-3">
          <div className="flex items-center justify-between">
            <p className="text-sm font-medium">Условие {index + 1}</p>
            <Button
              type="button"
              size="sm"
              variant="ghost"
              disabled={conditions.fields.length === 1}
              onClick={() => conditions.remove(index)}
            >
              Удалить условие
            </Button>
          </div>
          <LeafFields path={join(path, `conditions.${index}`)} />
        </div>
      ))}
      {message && <p className="text-xs text-destructive">{message}</p>}
      {groups.fields.map((field, index) => (
        <GroupEditor
          key={field.id}
          path={join(path, `groups.${index}`)}
          remainingDepth={remainingDepth - 1}
          onRemove={() => groups.remove(index)}
        />
      ))}
      <div className="flex flex-wrap gap-2">
        <Button
          type="button"
          size="sm"
          variant="outline"
          onClick={() => conditions.append({ ...emptyLeafValues })}
        >
          Добавить условие
        </Button>
        {remainingDepth > 0 && (
          <Button
            type="button"
            size="sm"
            variant="outline"
            onClick={() => groups.append(emptyGroupValues())}
          >
            Добавить группу
          </Button>
        )}
      </div>
    </div>
  )
}
