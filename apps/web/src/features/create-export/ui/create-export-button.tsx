import { useState } from 'react'
import { Button } from '@/shared/ui/button'
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/shared/ui/dialog'
import { ExportWizard } from './export-wizard'

interface CreateExportButtonProps {
  segmentId?: string
  disabled?: boolean
}

export function CreateExportButton({ segmentId, disabled }: CreateExportButtonProps) {
  const [open, setOpen] = useState(false)
  return (
    <>
      <Button disabled={disabled} onClick={() => setOpen(true)}>
        Выгрузить
      </Button>
      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Новая выгрузка</DialogTitle>
          </DialogHeader>
          {open && <ExportWizard segmentId={segmentId} onClose={() => setOpen(false)} />}
        </DialogContent>
      </Dialog>
    </>
  )
}
