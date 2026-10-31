import type { AccessAuditEntry } from '@/entities/access'
import { roleLabels, type Role } from '@/entities/user'
import type { SubjectDirectory } from '../api/use-subject-directory'

const roleLabel = (role: string) => (role in roleLabels ? roleLabels[role as Role] : role)

export function AuditActor({
  entry,
  directory,
}: {
  entry: AccessAuditEntry
  directory: SubjectDirectory
}) {
  return (
    <div className="flex flex-col">
      <span>{directory.userLabel(entry.actorId)}</span>
      <span className="text-xs text-muted-foreground">{roleLabel(entry.actorRole)}</span>
    </div>
  )
}
