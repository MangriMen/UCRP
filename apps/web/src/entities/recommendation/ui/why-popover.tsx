import { Button } from '@/shared/ui/button'
import {
  Popover,
  PopoverContent,
  PopoverDescription,
  PopoverHeader,
  PopoverTitle,
  PopoverTrigger,
} from '@/shared/ui/popover'
import { factorLabels } from '../model/labels'
import type { Recommendation } from '../model/types'

const formatShare = (value: number) => `${Math.round(value * 100)}%`

export function WhyPopover({ recommendation }: { recommendation: Recommendation }) {
  return (
    <Popover>
      <PopoverTrigger render={<Button variant="link" size="sm" />}>Почему?</PopoverTrigger>
      <PopoverContent align="start">
        <PopoverHeader>
          <PopoverTitle>Почему рекомендуем</PopoverTitle>
          <PopoverDescription>{recommendation.explanation}</PopoverDescription>
        </PopoverHeader>
        {recommendation.factors && recommendation.factors.length > 0 && (
          <ul className="flex flex-col gap-1">
            {recommendation.factors.map((factor) => (
              <li key={factor.factorCode} className="flex justify-between gap-2">
                <span>{factorLabels[factor.factorCode]}</span>
                <span className="text-muted-foreground">{formatShare(factor.contribution)}</span>
              </li>
            ))}
          </ul>
        )}
      </PopoverContent>
    </Popover>
  )
}
