import type { AccessPermission, SubjectPermissions } from '@/entities/access'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/shared/ui/card'

interface SimulationResultProps {
  result: SubjectPermissions
  current: string[] | undefined
  permissions: AccessPermission[]
}

function PermissionList({
  title,
  codes,
  titleOf,
}: {
  title: string
  codes: string[]
  titleOf: (code: string) => string
}) {
  return (
    <div className="flex flex-col gap-1">
      <p className="text-sm font-medium">{title}</p>
      {codes.length === 0 ? (
        <p className="text-sm text-muted-foreground">Нет</p>
      ) : (
        <ul className="list-disc pl-5 text-sm">
          {codes.map((code) => (
            <li key={code}>{titleOf(code)}</li>
          ))}
        </ul>
      )}
    </div>
  )
}

export function SimulationResult({ result, current, permissions }: SimulationResultProps) {
  const titleOf = (code: string) => permissions.find((p) => p.code === code)?.title ?? code
  const before = new Set(current ?? [])
  const after = new Set(result.permissions)
  const added = result.permissions.filter((code) => !before.has(code))
  const removed = [...before].filter((code) => !after.has(code))

  return (
    <Card>
      <CardHeader>
        <CardTitle>Что получит субъект</CardTitle>
        <CardDescription>
          {`После выдачи: ${result.permissions.length} прав. Грант не сохранён.`}
          {result.userGrantsIncluded && ' Учтены личные гранты пользователя.'}
        </CardDescription>
      </CardHeader>
      <CardContent className="grid gap-4 sm:grid-cols-2">
        <PermissionList title="Добавится" codes={added} titleOf={titleOf} />
        <PermissionList title="Пропадёт" codes={removed} titleOf={titleOf} />
      </CardContent>
    </Card>
  )
}
