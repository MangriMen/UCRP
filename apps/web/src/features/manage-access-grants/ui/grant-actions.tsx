import { useState } from 'react'
import { Link } from 'react-router'
import {
  availableTransitions,
  grantTransitionLabels,
  isGrantEditable,
  type AccessGrant,
  type GrantTransition,
} from '@/entities/access'
import { useAuth } from '@/entities/user'
import { Button } from '@/shared/ui/button'
import { GrantTransitionDialog } from './grant-transition-dialog'

export function GrantActions({ grant }: { grant: AccessGrant }) {
  const { can } = useAuth()
  const [transition, setTransition] = useState<GrantTransition | null>(null)
  if (!can('access.grant.write')) return null

  return (
    <div className="flex flex-wrap gap-2">
      {isGrantEditable(grant) && (
        <Button
          variant="outline"
          nativeButton={false}
          render={<Link to={`/access/grants/${grant.id}/edit`} />}
        >
          Изменить
        </Button>
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
