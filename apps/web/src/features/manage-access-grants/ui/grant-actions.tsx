import { useState } from 'react'
import {
  availableTransitions,
  grantTransitionLabels,
  isGrantEditable,
  type AccessGrant,
  type GrantTransition,
} from '@/entities/access'
import { useAuth } from '@/entities/user'
import { Button, ButtonLink } from '@/shared/ui/button'
import { GrantTransitionDialog } from './grant-transition-dialog'

export function GrantActions({ grant }: { grant: AccessGrant }) {
  const { can } = useAuth()
  const [transition, setTransition] = useState<GrantTransition | null>(null)
  if (!can('access.grant.write')) return null

  return (
    <div className="flex flex-wrap gap-2">
      {isGrantEditable(grant) && (
        <ButtonLink to={`/access/grants/${grant.id}/edit`} variant="outline">
          Изменить
        </ButtonLink>
      )}
      {availableTransitions(grant).map((next) => (
        <Button
          key={next}
          variant={next === 'revoke' ? 'destructive' : 'outline'}
          onClick={() => setTransition(next)}
        >
          {grantTransitionLabels[next]}
        </Button>
      ))}
      {transition && (
        <GrantTransitionDialog
          grant={grant}
          transition={transition}
          onClose={() => setTransition(null)}
        />
      )}
    </div>
  )
}
