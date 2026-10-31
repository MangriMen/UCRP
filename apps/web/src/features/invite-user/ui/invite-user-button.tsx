import { useState } from 'react'
import { Button } from '@/shared/ui/button'
import { InviteUserDialog } from './invite-user-dialog'

export function InviteUserButton() {
  const [open, setOpen] = useState(false)
  return (
    <>
      <Button onClick={() => setOpen(true)}>Пригласить пользователя</Button>
      {open && <InviteUserDialog open onOpenChange={setOpen} />}
    </>
  )
}
