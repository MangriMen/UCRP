import { useState } from 'react'
import {
  DownloadExportButton,
  exportChannelLabels,
  type Export,
  type ExportChannel,
} from '@/entities/export'
import { useSegments } from '@/entities/segment'
import { Alert, AlertDescription } from '@/shared/ui/alert'
import { Button } from '@/shared/ui/button'
import { DialogFooter } from '@/shared/ui/dialog'
import { ErrorBanner } from '@/shared/ui/error-banner'
import { FormField } from '@/shared/ui/form-field'
import { Input } from '@/shared/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/shared/ui/select'
import { useCreateExport } from '../api/use-create-export'
import { usePreviewExport } from '../api/use-preview-export'
import { exportErrorMessage } from '../model/errors'
import { CountsCard } from './counts-card'

const MAX_CONTROL_PERCENT = 50

const channelItems = Object.entries(exportChannelLabels).map(([value, label]) => ({
  value: value as ExportChannel,
  label,
}))

type Step = 'params' | 'audience' | 'confirm' | 'result'

interface ExportWizardProps {
  segmentId?: string
  onClose: () => void
}

export function ExportWizard({ segmentId: fixedSegmentId, onClose }: ExportWizardProps) {
  const segments = useSegments({ status: 'ready', pageSize: 100 })
  const create = useCreateExport()
  const [step, setStep] = useState<Step>('params')
  const [pickedSegmentId, setPickedSegmentId] = useState<string | null>(null)
  const [channel, setChannel] = useState<ExportChannel>('sms')
  const [controlPercent, setControlPercent] = useState(0)
  const [idempotencyKey, setIdempotencyKey] = useState(() => crypto.randomUUID())
  const [created, setCreated] = useState<Export | null>(null)

  const segmentId = fixedSegmentId ?? pickedSegmentId
  const preview = usePreviewExport(
    { segmentId: segmentId ?? '', channel, controlPercent },
    step === 'audience' && Boolean(segmentId),
  )
  const segmentItems = (segments.data?.items ?? []).map((s) => ({ value: s.id, label: s.name }))
  const controlValid =
    Number.isInteger(controlPercent) && controlPercent >= 0 && controlPercent <= MAX_CONTROL_PERCENT

  const goTo = (next: Step) => {
    create.reset()
    setStep(next)
  }

  const submit = () => {
    if (!segmentId) return
    create.mutate(
      { request: { segmentId, channel, format: 'csv', controlPercent }, idempotencyKey },
      {
        onSuccess: (item) => {
          setCreated(item)
          setStep('result')
        },
      },
    )
  }

  const changeParams = (apply: () => void) => {
    apply()
    setIdempotencyKey(crypto.randomUUID())
  }

  if (step === 'result' && created) {
    return (
      <div className="flex flex-col gap-4">
        <Alert>
          <AlertDescription>
            Выгрузка создана: {created.exportedCount} получателей, контрольная группа{' '}
            {created.controlCount ?? 0}.
          </AlertDescription>
        </Alert>
        <DialogFooter>
          {created.status === 'ready' && <DownloadExportButton exportId={created.id} />}
          <Button onClick={onClose}>Закрыть</Button>
        </DialogFooter>
      </div>
    )
  }

  const limitMessage = exportErrorMessage(create.error)

  return (
    <div className="flex flex-col gap-4">
      {step === 'params' && (
        <>
          {!fixedSegmentId && (
            <FormField label="Сегмент">
              {(field) => (
                <Select
                  items={segmentItems}
                  value={pickedSegmentId}
                  onValueChange={(value) => changeParams(() => setPickedSegmentId(value))}
                >
                  <SelectTrigger {...field} className="w-full">
                    <SelectValue placeholder="Выберите сегмент" />
                  </SelectTrigger>
                  <SelectContent>
                    {segmentItems.map((item) => (
                      <SelectItem key={item.value} value={item.value}>
                        {item.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              )}
            </FormField>
          )}
          <FormField label="Канал">
            {(field) => (
              <Select
                items={channelItems}
                value={channel}
                onValueChange={(value) =>
                  changeParams(() => value && setChannel(value as ExportChannel))
                }
              >
                <SelectTrigger {...field} className="w-full">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {channelItems.map((item) => (
                    <SelectItem key={item.value} value={item.value}>
                      {item.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            )}
          </FormField>
          <DialogFooter>
            <Button variant="outline" onClick={onClose}>
              Отмена
            </Button>
            <Button disabled={!segmentId} onClick={() => goTo('audience')}>
              Далее
            </Button>
          </DialogFooter>
        </>
      )}

      {step === 'audience' && (
        <>
          <FormField
            label="Контрольная группа, %"
            hint="От 0 до 50. Контрольная группа не попадает в файл."
            error={controlValid ? undefined : 'Введите целое число от 0 до 50'}
          >
            {(field) => (
              <Input
                {...field}
                type="number"
                min={0}
                max={MAX_CONTROL_PERCENT}
                value={controlPercent}
                onChange={(event) =>
                  changeParams(() => setControlPercent(event.target.valueAsNumber))
                }
              />
            )}
          </FormField>
          {preview.error && <ErrorBanner error={preview.error} />}
          {preview.data && controlValid && <CountsCard preview={preview.data} />}
          <DialogFooter>
            <Button variant="outline" onClick={() => goTo('params')}>
              Назад
            </Button>
            <Button
              disabled={!controlValid || !preview.data || preview.data.toSend === 0}
              onClick={() => goTo('confirm')}
            >
              Далее
            </Button>
          </DialogFooter>
        </>
      )}

      {step === 'confirm' && (
        <>
          <p className="text-sm">
            Будет создана выгрузка в канал «{exportChannelLabels[channel]}», формат CSV, контрольная
            группа {controlPercent}%. Отправлено будет {preview.data?.toSend ?? 0} получателей.
          </p>
          {limitMessage ? (
            <Alert variant="destructive">
              <AlertDescription>{limitMessage}</AlertDescription>
            </Alert>
          ) : (
            create.error && <ErrorBanner error={create.error} />
          )}
          <DialogFooter>
            <Button variant="outline" onClick={() => goTo('audience')}>
              Назад
            </Button>
            <Button disabled={create.isPending} onClick={submit}>
              Создать выгрузку
            </Button>
          </DialogFooter>
        </>
      )}
    </div>
  )
}
