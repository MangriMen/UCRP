import type { ExportPreview } from '@/entities/export'
import { Card, CardContent, CardHeader, CardTitle } from '@/shared/ui/card'

const rows = (preview: ExportPreview): [string, number][] => [
  ['В сегменте', preview.inSegment],
  ['С контактом для канала', preview.withContact],
  ['Исключены лимитом', preview.blockedByLimit],
  ['Контрольная группа', preview.control],
  ['Будет отправлено', preview.toSend],
]

export function CountsCard({ preview }: { preview: ExportPreview }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Аудитория выгрузки</CardTitle>
      </CardHeader>
      <CardContent>
        <dl className="grid grid-cols-[1fr_auto] gap-x-6 gap-y-1 text-sm">
          {rows(preview).map(([label, value]) => (
            <div key={label} className="contents">
              <dt className="text-muted-foreground">{label}</dt>
              <dd className="text-right font-medium">{value}</dd>
            </div>
          ))}
        </dl>
      </CardContent>
    </Card>
  )
}
