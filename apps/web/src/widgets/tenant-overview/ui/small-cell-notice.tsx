import { Info } from 'lucide-react'
import { Alert, AlertDescription, AlertTitle } from '@/shared/ui/alert'

interface SmallCellNoticeProps {
  threshold: number
}

export function SmallCellNotice({ threshold }: SmallCellNoticeProps) {
  return (
    <Alert>
      <Info />
      <AlertTitle>Часть значений скрыта</AlertTitle>
      <AlertDescription>
        Значения меньше {threshold} не показываются, чтобы нельзя было определить отдельного гостя.
      </AlertDescription>
    </Alert>
  )
}
