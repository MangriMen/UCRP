import type { ExportChannel, ExportStatus } from './types'

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

export const isExportActive = (status: ExportStatus) => status === 'queued' || status === 'building'
