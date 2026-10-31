import { Button } from '@/shared/ui/button'
import { useLogout } from '../api/use-logout'

export function LogoutButton() {
  const logout = useLogout()
  return (
    <Button variant="ghost" size="sm" onClick={() => logout.mutate()} disabled={logout.isPending}>
      Выйти
    </Button>
  )
}
