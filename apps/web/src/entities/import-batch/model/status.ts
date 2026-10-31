import type { ImportStatus, RejectedRowCode } from './types'

export const importStatusLabels: Record<ImportStatus, string> = {
  queued: 'В очереди',
  running: 'Выполняется',
  completed: 'Завершён',
  completed_with_errors: 'Завершён с ошибками',
  failed: 'Ошибка',
}

export const rejectedCodeLabels: Record<RejectedRowCode, string> = {
  bad_date: 'Некорректная дата',
  no_external_id: 'Нет внешнего идентификатора',
  bad_amount: 'Некорректная сумма',
  unknown_event_type: 'Неизвестный тип события',
  bad_phone: 'Некорректный телефон',
}

export const isImportActive = (status: ImportStatus) => status === 'queued' || status === 'running'
