import { useState } from 'react'
import type { Segment } from '@/entities/segment'
import { Button } from '@/shared/ui/button'
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/shared/ui/dialog'
import { RuleBuilderForm } from './rule-builder-form'

interface BuildSegmentButtonProps {
  segment?: Segment
}

export function BuildSegmentButton({ segment }: BuildSegmentButtonProps) {
  const [open, setOpen] = useState(false)
  return (
    <>
      <Button variant={segment ? 'outline' : 'default'} onClick={() => setOpen(true)}>
        {segment ? 'Изменить правило' : 'Новый сегмент'}
      </Button>
      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent className="max-h-[90vh] overflow-y-auto sm:max-w-2xl">
          <DialogHeader>
            <DialogTitle>{segment ? 'Новая версия сегмента' : 'Новый сегмент'}</DialogTitle>
          </DialogHeader>
          {open && <RuleBuilderForm segment={segment} onDone={() => setOpen(false)} />}
        </DialogContent>
      </Dialog>
    </>
  )
}
