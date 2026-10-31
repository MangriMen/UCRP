import { UploadIcon } from 'lucide-react'
import { useRef, useState, type DragEvent } from 'react'
import { cn } from '@/shared/lib/utils'
import { ACCEPTED_EXTENSIONS } from '../model/validate-file'

interface FileDropzoneProps {
  file: File | null
  onFile: (file: File) => void
}

export function FileDropzone({ file, onFile }: FileDropzoneProps) {
  const inputRef = useRef<HTMLInputElement>(null)
  const [isDragging, setIsDragging] = useState(false)

  const handleDrop = (event: DragEvent) => {
    event.preventDefault()
    setIsDragging(false)
    const dropped = event.dataTransfer.files[0]
    if (dropped) onFile(dropped)
  }

  return (
    <button
      type="button"
      onClick={() => inputRef.current?.click()}
      onDragOver={(event) => {
        event.preventDefault()
        setIsDragging(true)
      }}
      onDragLeave={() => setIsDragging(false)}
      onDrop={handleDrop}
      className={cn(
        'flex w-full flex-col items-center gap-2 rounded-lg border-2 border-dashed p-8 text-sm text-muted-foreground transition-colors hover:bg-muted/50',
        isDragging && 'border-primary bg-muted/50',
      )}
    >
      <UploadIcon className="size-6" />
      {file ? (
        <span className="font-medium text-foreground">{file.name}</span>
      ) : (
        <span>Перетащите файл сюда или нажмите, чтобы выбрать</span>
      )}
      <span className="text-xs">CSV или Excel, до 10 МБ</span>
      <input
        ref={inputRef}
        type="file"
        accept={ACCEPTED_EXTENSIONS.join(',')}
        className="sr-only"
        aria-label="Файл для импорта"
        onChange={(event) => {
          const selected = event.target.files?.[0]
          if (selected) onFile(selected)
          event.target.value = ''
        }}
      />
    </button>
  )
}
