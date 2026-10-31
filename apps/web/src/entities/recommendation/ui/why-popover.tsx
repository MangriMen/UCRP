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
import { isLowData, scoredFactors } from '../model/factors'
import type { Recommendation } from '../model/types'

const formatShare = (value: number) => `${Math.round(value * 100)}%`

export function WhyPopover({ recommendation }: { recommendation: Recommendation }) {
  const factors = scoredFactors(recommendation)

  return (
    <Popover>
      <PopoverTrigger render={<Button variant="link" size="sm" />}>Почему?</PopoverTrigger>
      <PopoverContent align="start">
        <PopoverHeader>
          <PopoverTitle>Почему рекомендуем</PopoverTitle>
          <PopoverDescription>{recommendation.explanation}</PopoverDescription>
        </PopoverHeader>
        {isLowData(recommendation) && (
          <p className="text-sm text-muted-foreground">
            Истории гостя мало, поэтому рекомендация основана на популярности и приоритете
            арендатора.
          </p>
        )}
        {factors.length > 0 && (
          <ul className="flex flex-col gap-1">
            {factors.map((factor) => (
              <li key={factor.factorCode} className="flex flex-col">
                <span className="flex justify-between gap-2">
                  <span>{factorLabels[factor.factorCode]}</span>
                  <span className="tabular-nums">{formatShare(factor.contribution)}</span>
                </span>
                <span className="text-xs text-muted-foreground">
                  Значение {formatShare(factor.rawValue)} · вес {formatShare(factor.weight)}
                </span>
              </li>
            ))}
          </ul>
        )}
        <p className="text-xs text-muted-foreground">
          Оценка {recommendation.score.toFixed(2)}
          {recommendation.algorithmVersion && ` · алгоритм ${recommendation.algorithmVersion}`}
        </p>
      </PopoverContent>
    </Popover>
  )
}
