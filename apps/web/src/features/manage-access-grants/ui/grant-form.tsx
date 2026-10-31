import { useState, type FormEvent, type ReactNode } from 'react'
import {
  presetLabel,
  useAccessPermissions,
  useAccessPresets,
  useEffectiveAccess,
  type AccessGrant,
  type AccessSubject,
} from '@/entities/access'
import { useAuth } from '@/entities/user'
import { Button } from '@/shared/ui/button'
import { CheckboxField } from '@/shared/ui/checkbox-field'
import { ErrorBanner } from '@/shared/ui/error-banner'
import { FormField } from '@/shared/ui/form-field'
import { Input } from '@/shared/ui/input'
import { OptionSelect } from '@/shared/ui/option-select'
import { Skeleton } from '@/shared/ui/skeleton'
import { Textarea } from '@/shared/ui/textarea'
import { useCreateGrant, useSimulateAccess, useUpdateGrant } from '../api/use-grant-mutations'
import { useSubjectDirectory } from '../api/use-subject-directory'
import {
  emptyGrantForm,
  grantFormFromGrant,
  toGrantPatch,
  toGrantRequest,
  validateGrantForm,
  type GrantFormErrors,
  type GrantFormValues,
} from '../model/grant-form'
import {
  fromSubject,
  grantableRoleOf,
  subjectKindLabels,
  tenantScopeOf,
  toSubject,
  type SubjectValue,
} from '../model/subject'
import { GrantError } from './grant-error'
import { PermissionChecklist } from './permission-checklist'
import { ScopesEditor } from './scopes-editor'
import { SimulationResult } from './simulation-result'
import { SubjectPicker } from './subject-picker'

const NO_PRESET = '__none__'

const statusItems = [
  { value: 'active', label: 'Сразу действует' },
  { value: 'draft', label: 'Черновик' },
]

interface GrantFormProps {
  grant?: AccessGrant
  initialSubject?: AccessSubject | null
  onSaved: (grant: AccessGrant) => void
  onCancel: () => void
}

function Section({
  title,
  error,
  children,
}: {
  title: string
  error?: string
  children: ReactNode
}) {
  return (
    <section className="flex flex-col gap-3">
      <h2 className="text-base font-semibold">{title}</h2>
      {children}
      {error && <p className="text-sm text-destructive">{error}</p>}
    </section>
  )
}

const DEFAULT_SUBJECT: SubjectValue = { kind: 'tenant', id: '' }

export function GrantForm({ grant, initialSubject, onSaved, onCancel }: GrantFormProps) {
  const { can } = useAuth()
  const directory = useSubjectDirectory()
  const permissions = useAccessPermissions()
  const presets = useAccessPresets()
  const create = useCreateGrant()
  const update = useUpdateGrant()
  const simulate = useSimulateAccess()
  const [chosenSubject, setChosenSubject] = useState<SubjectValue | null>(null)
  const [values, setValues] = useState<GrantFormValues>(() =>
    grant ? grantFormFromGrant(grant) : emptyGrantForm(),
  )
  const [errors, setErrors] = useState<GrantFormErrors>({})

  const fixedSubject = grant ?? initialSubject
  const subject =
    chosenSubject ?? (fixedSubject ? fromSubject(fixedSubject, directory.users) : DEFAULT_SUBJECT)
  const tenantScope = tenantScopeOf(subject, directory.users)
  const role = grantableRoleOf(subject.kind)
  const isManager = subject.kind === 'manager'
  const mutation = grant ? update : create
  const simulatedSubject = simulate.data ? toSubject(subject) : null
  const current = useEffectiveAccess(simulatedSubject)

  const change = (patch: Partial<GrantFormValues>) => {
    simulate.reset()
    setValues((previous) => ({ ...previous, ...patch }))
  }

  const changeSubject = (next: SubjectValue) => {
    simulate.reset()
    setChosenSubject(next)
  }

  const validated = () => {
    const found = validateGrantForm(values, subject, tenantScope)
    setErrors(found)
    return Object.keys(found).length === 0
  }

  const submit = (event: FormEvent) => {
    event.preventDefault()
    if (!validated()) return
    if (grant) {
      update.mutate(
        { id: grant.id, body: toGrantPatch(values, subject, tenantScope) },
        { onSuccess: onSaved },
      )
      return
    }
    const body = toGrantRequest(values, subject, tenantScope)
    if (body) create.mutate(body, { onSuccess: onSaved })
  }

  const runSimulation = () => {
    if (!validated()) return
    const body = toGrantRequest(values, subject, tenantScope)
    if (body) simulate.mutate({ ...body, status: 'active' })
  }

  if (permissions.isLoading || presets.isLoading) return <Skeleton className="h-64 w-full" />
  if (permissions.error) return <ErrorBanner error={permissions.error} />

  const presetItems = [
    { value: NO_PRESET, label: 'Без пресета' },
    ...(presets.data ?? [])
      .filter((preset) => preset.available || preset.code === values.presetCode)
      .map((preset) => ({ value: preset.code, label: presetLabel(preset.code) })),
  ]
  const unavailablePresets = (presets.data ?? []).filter((preset) => !preset.available)

  const choosePreset = (code: string) => {
    const preset = presets.data?.find((candidate) => candidate.code === code)
    change(preset ? { presetCode: code, items: preset.items } : { presetCode: '' })
  }

  return (
    <form onSubmit={submit} noValidate className="flex max-w-3xl flex-col gap-8">
      <Section title="Субъект" error={errors.subject}>
        {grant ? (
          <p className="text-sm">
            {`${subjectKindLabels[subject.kind]}: ${directory.subjectLabel(grant)}`}
          </p>
        ) : (
          <SubjectPicker
            value={subject}
            onChange={changeSubject}
            directory={directory}
            excludeSelf
            invalid={Boolean(errors.subject)}
          />
        )}
        {isManager && (
          <p className="text-sm text-muted-foreground">
            Управляющему можно выдать только права на личные данные гостей, без пресета и
            ограничений. Выдать грант самому себе нельзя.
          </p>
        )}
      </Section>

      {!isManager && (
        <Section title="Пресет">
          <OptionSelect
            ariaLabel="Пресет"
            className="w-72"
            items={presetItems}
            value={values.presetCode || NO_PRESET}
            onValueChange={(code) => choosePreset(code === NO_PRESET ? '' : code)}
          />
          {unavailablePresets.length > 0 && (
            <p className="text-xs text-muted-foreground">
              {`Недоступны от вашего имени: ${unavailablePresets.map((preset) => presetLabel(preset.code)).join(', ')}`}
            </p>
          )}
        </Section>
      )}

      <Section title="Права" error={errors.items}>
        <PermissionChecklist
          permissions={permissions.data ?? []}
          role={role}
          items={values.items}
          onChange={(items) => change({ items, presetCode: '' })}
        />
      </Section>

      {!isManager && (
        <Section title="Ограничения" error={errors.scopes}>
          <p className="text-sm text-muted-foreground">
            {tenantScope
              ? `Область арендатора добавляется автоматически: ${tenantScope === 'self' ? 'свой арендатор каждого пользователя' : directory.subjectLabel({ subjectType: 'tenant', subjectId: tenantScope })}.`
              : 'Область арендатора появится после выбора субъекта.'}
          </p>
          <ScopesEditor scopes={values.scopes} onChange={(scopes) => change({ scopes })} />
        </Section>
      )}

      <Section title="Срок действия">
        <div className="flex flex-wrap items-start gap-4">
          <FormField label="Начало" error={errors.validFrom}>
            {(field) => (
              <Input
                {...field}
                type="datetime-local"
                className="w-56"
                value={values.validFrom}
                onChange={(event) => change({ validFrom: event.target.value })}
              />
            )}
          </FormField>
          <FormField label="Окончание" error={errors.validTo}>
            {(field) => (
              <Input
                {...field}
                type="datetime-local"
                className="w-56"
                disabled={values.indefinite}
                value={values.indefinite ? '' : values.validTo}
                onChange={(event) => change({ validTo: event.target.value })}
              />
            )}
          </FormField>
        </div>
        <CheckboxField
          label="Бессрочно"
          checked={values.indefinite}
          onCheckedChange={(indefinite) => change({ indefinite })}
        />
      </Section>

      {!grant && (
        <Section title="Статус">
          <OptionSelect
            ariaLabel="Статус"
            className="w-56"
            items={statusItems}
            value={values.status}
            onValueChange={(status) => change({ status: status === 'draft' ? 'draft' : 'active' })}
          />
        </Section>
      )}

      <FormField label="Причина" error={errors.reason}>
        {(field) => (
          <Textarea
            {...field}
            value={values.reason}
            onChange={(event) => change({ reason: event.target.value })}
            placeholder="Например, договор о пилоте от 01.11.2026"
          />
        )}
      </FormField>

      {mutation.error && <GrantError error={mutation.error} />}
      {simulate.error && <GrantError error={simulate.error} />}
      {simulate.data && (
        <SimulationResult
          result={simulate.data}
          current={current.data?.permissions}
          permissions={permissions.data ?? []}
        />
      )}

      <div className="flex flex-wrap gap-2">
        <Button type="submit" disabled={mutation.isPending}>
          {grant ? 'Сохранить' : 'Выдать доступ'}
        </Button>
        {!grant && can('access.simulate') && (
          <Button
            type="button"
            variant="outline"
            disabled={simulate.isPending}
            onClick={runSimulation}
          >
            Проверить доступ
          </Button>
        )}
        <Button type="button" variant="ghost" onClick={onCancel}>
          Отмена
        </Button>
      </div>
    </form>
  )
}
