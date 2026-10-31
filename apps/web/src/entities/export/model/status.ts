import type { ExportChannel, ExportFormat, ExportStatus } from './types'

export const exportStatusLabels: Record<ExportStatus, string> = {
  queued: 'В очереди',
  building: 'Формируется',
  ready: 'Готов',
  failed: 'Ошибка',
}

export const exportChannelLabels: Record<ExportChannel, string> = {
  sms: 'SMS',
  push: 'Push',
  email: 'Email',
}

export const exportFormatLabels: Record<ExportFormat, string> = {
  csv: 'CSV',
  api: 'API',
}

export const isExportActive = (status: ExportStatus) => status === 'queued' || status === 'building'
