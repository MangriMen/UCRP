export const warningReasonLabels: Record<string, string> = {
  rejected_share_above_10pct: 'Отклонено более 10% строк',
  data_older_than_7_days: 'Данные старше 7 дней',
}

export const warningReasonLabel = (reason: string) => warningReasonLabels[reason] ?? reason
