import type { SegmentRule, SegmentRuleGroup, SegmentRuleNode } from './types'

export const isRuleGroup = (node: SegmentRuleNode): node is SegmentRuleGroup => 'conditions' in node

export function collectLeaves(node: SegmentRuleNode): SegmentRule[] {
  return isRuleGroup(node) ? node.conditions.flatMap(collectLeaves) : [node]
}

const countOf = (ids: string[] | undefined, one: string, many: string) =>
  ids && ids.length > 0 ? `${one} ${ids.length} ${many}` : null

const MINOR_PER_MAJOR = 100

export function describeLeaf(rule: SegmentRule): string {
  const parts = [
    countOf(rule.visitedTenantIds, 'был у', 'арендаторов'),
    countOf(rule.excludedTenantIds, 'не был у', 'арендаторов'),
    countOf(rule.visitedCategoryIds, 'был в', 'категориях'),
    countOf(rule.excludedCategoryIds, 'не был в', 'категориях'),
    rule.purchasedZones && rule.purchasedZones.length > 0
      ? `покупал на ${rule.purchasedZones.length} этажах`
      : null,
    rule.requireContact
      ? `контакт: ${rule.requireContact === 'phone' ? 'телефон' : 'email'}`
      : null,
    rule.minVisits != null ? `от ${rule.minVisits} визитов` : null,
    rule.minAmountMinor != null ? `от ${rule.minAmountMinor / MINOR_PER_MAJOR} ₽` : null,
    `за ${rule.periodDays} дн.`,
  ]
  return parts.filter((part) => part !== null).join(', ')
}

export const groupLabel = (group: SegmentRuleGroup) =>
  group.operator === 'and' ? 'Все условия:' : 'Любое из условий:'
