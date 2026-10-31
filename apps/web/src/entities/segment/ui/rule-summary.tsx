import { describeLeaf, groupLabel, isRuleGroup } from '../model/rule'
import type { SegmentRuleNode } from '../model/types'

interface RuleSummaryProps {
  rule: SegmentRuleNode
}

export function RuleSummary({ rule }: RuleSummaryProps) {
  if (!isRuleGroup(rule)) return <p>{describeLeaf(rule)}</p>

  return (
    <div className="flex flex-col gap-1">
      <p className="font-medium">{groupLabel(rule)}</p>
      <ul className="ml-4 flex list-disc flex-col gap-1">
        {rule.conditions.map((condition, index) => (
          <li key={index}>
            <RuleSummary rule={condition} />
          </li>
        ))}
      </ul>
    </div>
  )
}
