import { GuestStatusBadge, identifierTypeLabels, type GuestProfile } from '@/entities/guest'
import type { MergeCandidateDetail } from '@/entities/merge-candidate'
import { formatDateTime } from '@/shared/lib/format'
import { Badge } from '@/shared/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/shared/ui/card'

interface MergeCompareProps {
  detail: MergeCandidateDetail
}

function ProfileColumn({ title, profile }: { title: string; profile: GuestProfile }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex flex-wrap items-center gap-2">
          {title}: {profile.displayName ?? 'Без имени'}
          <GuestStatusBadge status={profile.status} />
        </CardTitle>
      </CardHeader>
      <CardContent className="flex flex-col gap-4 text-sm">
        <dl className="grid grid-cols-2 gap-3">
          <div>
            <dt className="text-xs text-muted-foreground">ID</dt>
            <dd className="break-all">{profile.id}</dd>
          </div>
          <div>
            <dt className="text-xs text-muted-foreground">Событий</dt>
            <dd>{profile.eventCount ?? 0}</dd>
          </div>
          <div>
            <dt className="text-xs text-muted-foreground">Арендаторов</dt>
            <dd>{profile.tenantCount}</dd>
          </div>
          <div>
            <dt className="text-xs text-muted-foreground">Последняя активность</dt>
            <dd>{profile.lastSeenAt ? formatDateTime(profile.lastSeenAt) : '—'}</dd>
          </div>
        </dl>
        <div>
          <p className="mb-1 text-xs text-muted-foreground">Идентификаторы</p>
          {profile.identifiers.length === 0 ? (
            <p className="text-muted-foreground">Нет</p>
          ) : (
            <ul className="flex flex-col gap-1">
              {profile.identifiers.map((identifier) => (
                <li key={identifier.id}>
                  {identifierTypeLabels[identifier.idType]}: {identifier.valueNormalized}
                </li>
              ))}
            </ul>
          )}
        </div>
      </CardContent>
    </Card>
  )
}

export function MergeCompare({ detail }: MergeCompareProps) {
  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-wrap items-center gap-2 text-sm">
        <span className="text-muted-foreground">Общие ключи:</span>
        {detail.sharedKeys.length === 0 && <span>нет</span>}
        {detail.sharedKeys.map((key) => (
          <Badge key={`${key.idType}:${key.valueNormalized}`} variant="secondary">
            {key.idType ? identifierTypeLabels[key.idType] : 'Ключ'}: {key.valueNormalized}
          </Badge>
        ))}
        {detail.differingFields && detail.differingFields.length > 0 && (
          <>
            <span className="ml-2 text-muted-foreground">Различия:</span>
            {detail.differingFields.map((field) => (
              <Badge key={field} variant="outline">
                {field}
              </Badge>
            ))}
          </>
        )}
      </div>
      <div className="grid gap-4 md:grid-cols-2">
        <ProfileColumn title="Профиль A" profile={detail.guestA} />
        <ProfileColumn title="Профиль B" profile={detail.guestB} />
      </div>
    </div>
  )
}
