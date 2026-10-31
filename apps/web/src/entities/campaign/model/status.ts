import type { CampaignStatus } from './types'

export const campaignStatusLabels: Record<CampaignStatus, string> = {
  draft: 'Черновик',
  active: 'Идёт',
  finished: 'Завершена',
  archived: 'В архиве',
}
