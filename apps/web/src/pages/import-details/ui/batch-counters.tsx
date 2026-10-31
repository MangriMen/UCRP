import type { ImportBatch } from '@/entities/import-batch'
import { Card, CardContent, CardHeader, CardTitle } from '@/shared/ui/card'

const counters = [
  { label: 'Всего строк', value: (b: ImportBatch) => b.totalRows },
  { label: 'Принято', value: (b: ImportBatch) => b.acceptedRows },
  { label: 'Отклонено', value: (b: ImportBatch) => b.rejectedRows },
  { label: 'Дубликаты', value: (b: ImportBatch) => b.duplicateRows },
]

export function BatchCounters({ batch }: { batch: ImportBatch }) {
  return (
    <div className="grid grid-cols-2 gap-4 lg:grid-cols-4">
      {counters.map((counter) => (
        <Card key={counter.label}>
          <CardHeader>
            <CardTitle className="text-sm font-normal text-muted-foreground">
              {counter.label}
            </CardTitle>
          </CardHeader>
          <CardContent className="text-2xl font-semibold">{counter.value(batch)}</CardContent>
        </Card>
      ))}
    </div>
  )
}
