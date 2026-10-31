import { ApiError } from '@/shared/api/client'

export function exportErrorMessage(error: unknown): string | null {
  if (!(error instanceof ApiError)) return null
  if (error.body.code === 'contact_limit_exceeded') {
    return 'Все получатели исключены лимитом контактов. Измените канал или смягчите лимиты в настройках.'
  }
  if (error.body.code === 'segment_not_ready') {
    return 'Сегмент ещё не рассчитан. Дождитесь окончания расчёта и повторите.'
  }
  return null
}
