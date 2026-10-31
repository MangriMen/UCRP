import {
  GuestStatusBadge,
  identifierTypeLabels,
  type GuestProfile,
  type GuestTenantSummary,
} from '@/entities/guest'
import { formatDateTime, formatMoney } from '@/shared/lib/format'
import { Badge } from '@/shared/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/shared/ui/card'
import { DataTable, type Column } from '@/shared/ui/data-table'

const tenantColumns: Column<GuestTenantSummary>[] = [
  { id: 'name', header: 'Арендатор', cell: (t) => t.tenantName },
  { id: 'category', header: 'Категория', cell: (t) => t.categoryName ?? '—' },
  { id: 'floor', header: 'Этаж', cell: (t) => t.floor ?? '—', className: 'text-right' },
  { id: 'events', header: 'Событий', cell: (t) => t.eventCount, className: 'text-right' },
  {
    id: 'lastVisit',
    header: 'Последний визит',
    cell: (t) => (t.lastVisitAt ? formatDateTime(t.lastVisitAt) : '—'),
  },
]

function Stat({ label, value }: { label: string; value: string | number }) {
  return (
    <div>
      <dt className="text-xs text-muted-foreground">{label}</dt>
      <dd className="text-sm font-medium">{value}</dd>
    </div>
  )
}

export function ProfileCard({ profile }: { profile: GuestProfile }) {
  return (
    <div className="flex flex-col gap-6">
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            {profile.displayName ?? 'Без имени'}
            <GuestStatusBadge status={profile.status} />
          </CardTitle>
        </CardHeader>
        <CardContent>
          <dl className="grid grid-cols-2 gap-4 sm:grid-cols-4">
            <Stat label="Арендаторов" value={profile.tenantCount} />
            <Stat label="Событий" value={profile.eventCount ?? 0} />
            <Stat
              label="Первое появление"
              value={profile.firstSeenAt ? formatDateTime(profile.firstSeenAt) : '—'}
            />
            <Stat
              label="Последняя активность"
              value={profile.lastSeenAt ? formatDateTime(profile.lastSeenAt) : '—'}
            />
            {profile.totalAmountMinor != null && (
              <Stat label="Сумма покупок" value={formatMoney(profile.totalAmountMinor)} />
            )}
          </dl>
        </CardContent>
      </Card>

      <section className="flex flex-col gap-3">
        <h2 className="text-lg font-semibold">Идентификаторы</h2>
        {profile.identifiers.length === 0 ? (
          <p className="text-sm text-muted-foreground">Идентификаторов нет</p>
        ) : (
          <ul className="flex flex-col gap-1 text-sm">
            {profile.identifiers.map((identifier) => (
              <li key={identifier.id}>
                <span className="text-muted-foreground">
                  {identifierTypeLabels[identifier.idType]}:{' '}
                </span>
                {identifier.valueNormalized}
              </li>
            ))}
          </ul>
        )}
      </section>

      <section className="flex flex-col gap-3">
        <h2 className="text-lg font-semibold">Сегменты</h2>
        {profile.segments.length === 0 ? (
          <p className="text-sm text-muted-foreground">Гость не входит в сегменты</p>
        ) : (
          <div className="flex flex-wrap gap-2">
            {profile.segments.map((segment) => (
              <Badge key={segment.segmentId} variant="secondary">
                {segment.name}
              </Badge>
            ))}
          </div>
        )}
      </section>

      <section className="flex flex-col gap-3">
        <h2 className="text-lg font-semibold">Активность по арендаторам</h2>
        <DataTable
          columns={tenantColumns}
          rows={profile.tenants}
          getRowId={(t) => t.tenantId}
          emptyTitle="Активности по арендаторам нет"
        />
      </section>
    </div>
  )
}
