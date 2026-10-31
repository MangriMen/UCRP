import { Link } from 'react-router'
import { Button } from '@/shared/ui/button'
import { EmptyState } from '@/shared/ui/empty-state'

export function NotFoundPage() {
  return (
    <EmptyState
      title="Страница не найдена"
      description="Такого адреса нет или раздел ещё не готов."
      action={<Button render={<Link to="/" />}>На главную</Button>}
    />
  )
}
