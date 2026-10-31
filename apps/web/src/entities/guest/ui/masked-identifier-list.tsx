import { identifierTypeLabels } from '../model/labels'
import type { MaskedGuestIdentifier } from '../model/types'

const VISIBLE_IDENTIFIERS = 2

export function MaskedIdentifierList({ identifiers }: { identifiers?: MaskedGuestIdentifier[] }) {
  if (!identifiers?.length) return <span className="text-muted-foreground">—</span>
  const hiddenCount = identifiers.length - VISIBLE_IDENTIFIERS
  return (
    <div className="flex flex-col gap-0.5">
      {identifiers.slice(0, VISIBLE_IDENTIFIERS).map((identifier, index) => (
        <span key={`${identifier.idType}-${index}`} className="whitespace-nowrap">
          <span className="text-muted-foreground">{identifierTypeLabels[identifier.idType]}:</span>{' '}
          <span className="font-mono">{identifier.valueMasked}</span>
        </span>
      ))}
      {hiddenCount > 0 && <span className="text-xs text-muted-foreground">ещё {hiddenCount}</span>}
    </div>
  )
}
