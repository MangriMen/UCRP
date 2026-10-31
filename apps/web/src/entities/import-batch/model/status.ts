import type { ImportStatus, RejectedRowCode } from './types'

export const importStatusLabels: Record<ImportStatus, string> = {
  queued: 'В очереди',
  running: 'Выполняется',
  completed: 'Завершён',
  completed_with_errors: 'Завершён с ошибками',
  failed: 'Ошибка',
}

export const rejectedCodeLabels: Record<RejectedRowCode, string> = {
  invalid_date: 'Некорректная дата',
  no_external_id: 'Нет внешнего идентификатора',
  bad_external_id: 'Некорректный внешний идентификатор',
  invalid_amount: 'Некорректная сумма',
  unknown_event_type: 'Неизвестный тип события',
  invalid_phone: 'Некорректный телефон',
  invalid_tenant: 'Неизвестный арендатор',
  no_tenant: 'Арендатор не указан',
  bad_email: 'Некорректный email',
  internal: 'Внутренняя ошибка обработки строки',
}

export const isImportActive = (status: ImportStatus) => status === 'queued' || status === 'running'
