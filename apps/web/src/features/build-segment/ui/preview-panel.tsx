import { useMemo } from 'react'
import type { SegmentRule } from '@/entities/segment'
import { useDebouncedValue } from '@/shared/lib/hooks/use-debounced-value'
import { formatDateTime } from '@/shared/lib/format'
import { Card, CardContent, CardHeader, CardTitle } from '@/shared/ui/card'
import { ErrorBanner } from '@/shared/ui/error-banner'
import { usePreviewSegment } from '../api/use-preview-segment'

const PREVIEW_DEBOUNCE_MS = 600

interface PreviewPanelProps {
  rule: SegmentRule | null
}

export function PreviewPanel({ rule }: PreviewPanelProps) {
  const ruleKey = useDebouncedValue(JSON.stringify(rule), PREVIEW_DEBOUNCE_MS)
  const debouncedRule = useMemo<SegmentRule | null>(() => JSON.parse(ruleKey), [ruleKey])
  const preview = usePreviewSegment(debouncedRule)
  const isStale = JSON.stringify(rule) !== ruleKey

  return (
    <Card>
      <CardHeader>
        <CardTitle>Предварительная аудитория</CardTitle>
      </CardHeader>
      <CardContent className="flex flex-col gap-3">
        {rule === null && (
          <p className="text-sm text-muted-foreground">Заполните условия корректно</p>
        )}
        {rule !== null && preview.error && <ErrorBanner error={preview.error} />}
        {rule !== null && !preview.error && !preview.data && (
          <p className="text-sm text-muted-foreground">Считаем…</p>
        )}
        {rule !== null && preview.data && (
          <div className={isStale || preview.isFetching ? 'opacity-60' : undefined}>
            <p className="text-2xl font-semibold">{preview.data.size}</p>
            <p className="text-xs text-muted-foreground">гостей подходит под правило</p>
            <ul className="mt-3 flex flex-col gap-1 text-sm">
              {preview.data.examples.map((guest) => (
                <li key={guest.guestId} className="flex justify-between gap-2">
                  <span>{guest.displayName ?? guest.guestId}</span>
                  <span className="text-muted-foreground">
                    {guest.lastSeenAt ? formatDateTime(guest.lastSeenAt) : '—'}
                  </span>
                </li>
              ))}
            </ul>
          </div>
        )}
      </CardContent>
    </Card>
  )
}
