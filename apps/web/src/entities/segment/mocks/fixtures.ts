import type { Segment, SegmentRuleNode, SegmentTemplate, SegmentVersion } from '../model/types'

export const mockSegments: Segment[] = [
  {
    id: 'segment-0001',
    suppressed: false,
    name: 'Спящий сосед',
    description: 'Гости, которые давно не заходили к соседям по этажу',
    templateCode: 'sleeping_neighbor',
    isTemplate: false,
    currentVersionNo: 2,
    status: 'ready',
    memberCount: 120,
    computedAt: '2026-10-30T06:00:00Z',
    rule: { visitedTenantIds: ['tenant-0001'], excludedTenantIds: ['tenant-0002'], periodDays: 90 },
    ruleText: 'Был в «Ленте» и не был в «Заре» за 90 дней',
    createdAt: '2026-09-01T09:00:00Z',
    updatedAt: '2026-10-30T06:00:00Z',
  },
  {
    id: 'segment-0002',
    suppressed: false,
    name: 'Высокий чек',
    description: null,
    templateCode: 'high_value',
    isTemplate: false,
    currentVersionNo: 1,
    status: 'ready',
    memberCount: 37,
    computedAt: '2026-10-29T06:00:00Z',
    rule: { periodDays: 180, minAmountMinor: 500000 },
    ruleText: 'Потратил от 5 000 ₽ за 180 дней',
    createdAt: '2026-09-15T09:00:00Z',
    updatedAt: '2026-10-29T06:00:00Z',
  },
  {
    id: 'segment-0003',
    suppressed: false,
    name: 'Риск оттока',
    description: 'Не возвращались более 60 дней',
    templateCode: 'churn_risk',
    isTemplate: false,
    currentVersionNo: 1,
    status: 'pending',
    memberCount: 0,
    computedAt: null,
    rule: { periodDays: 60, minVisits: 2 },
    ruleText: 'Минимум 2 визита, последний более 60 дней назад',
    createdAt: '2026-10-31T08:00:00Z',
    updatedAt: '2026-10-31T08:00:00Z',
  },
  {
    id: 'segment-0004',
    suppressed: false,
    name: 'Новички с телефоном',
    description: null,
    templateCode: null,
    isTemplate: false,
    currentVersionNo: 1,
    status: 'failed',
    memberCount: 0,
    computedAt: null,
    rule: { periodDays: 30, requireContact: 'phone' },
    ruleText: 'Новые гости за 30 дней с телефоном',
    createdAt: '2026-10-20T09:00:00Z',
    updatedAt: '2026-10-20T09:30:00Z',
  },
  {
    id: 'segment-0005',
    suppressed: false,
    name: 'Лояльные или крупные',
    description: 'Часто ходят к соседям либо тратят много',
    templateCode: null,
    isTemplate: false,
    currentVersionNo: 1,
    status: 'ready',
    memberCount: 64,
    computedAt: '2026-10-30T06:00:00Z',
    rule: {
      operator: 'and',
      conditions: [
        { periodDays: 90, requireContact: 'phone' },
        {
          operator: 'or',
          conditions: [
            { periodDays: 90, minVisits: 5 },
            { periodDays: 180, minAmountMinor: 500000 },
          ],
        },
      ],
    },
    ruleText: 'С телефоном и (от 5 визитов за 90 дней или от 5 000 ₽ за 180 дней)',
    createdAt: '2026-10-25T09:00:00Z',
    updatedAt: '2026-10-30T06:00:00Z',
  },
]

export function mockSegmentVersions(segment: Segment): SegmentVersion[] {
  return Array.from({ length: segment.currentVersionNo }, (_, index) => {
    const versionNo = segment.currentVersionNo - index
    const isCurrent = versionNo === segment.currentVersionNo
    const rule: SegmentRuleNode = isCurrent
      ? (segment.rule ?? { periodDays: 90 })
      : { periodDays: 30 }
    return {
      id: `${segment.id}-v${versionNo}`,
      segmentId: segment.id,
      versionNo,
      suppressed: false,
      rule,
      ruleText: isCurrent ? segment.ruleText : 'Предыдущая формулировка правила',
      status: isCurrent ? segment.status : 'ready',
      memberCount: isCurrent ? segment.memberCount : Math.round((segment.memberCount ?? 0) * 0.8),
      computedAt: isCurrent ? segment.computedAt : '2026-10-01T06:00:00Z',
      createdAt: isCurrent ? segment.updatedAt : segment.createdAt,
    }
  })
}

export const mockSegmentTemplates: SegmentTemplate[] = [
  {
    code: 'sleeping_neighbor',
    name: 'Спящий сосед',
    description: 'Был у одного арендатора и не заходил к другому',
    rule: { periodDays: 90 },
    defaultPeriodDays: 90,
  },
  {
    code: 'single_store',
    name: 'Один магазин',
    description: 'Покупает только в одном магазине',
    rule: { periodDays: 180 },
    defaultPeriodDays: 180,
  },
  {
    code: 'churn_risk',
    name: 'Риск оттока',
    description: 'Давно не возвращался',
    rule: { periodDays: 60, minVisits: 2 },
    defaultPeriodDays: 60,
  },
  {
    code: 'newcomer',
    name: 'Новичок',
    description: 'Первый визит недавно',
    rule: { periodDays: 30 },
    defaultPeriodDays: 30,
  },
  {
    code: 'high_value',
    name: 'Высокий чек',
    description: 'Крупные покупки за период или частые визиты',
    rule: {
      operator: 'or',
      conditions: [
        { periodDays: 180, minAmountMinor: 500000 },
        { periodDays: 90, minVisits: 5 },
      ],
    },
    defaultPeriodDays: 180,
  },
]
