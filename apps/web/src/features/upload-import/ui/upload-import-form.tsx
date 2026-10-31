import { useState, type FormEvent } from 'react'
import { useDataSourceOptions } from '@/entities/data-source'
import { Alert, AlertDescription } from '@/shared/ui/alert'
import { Button } from '@/shared/ui/button'
import { ErrorBanner } from '@/shared/ui/error-banner'
import { FormField } from '@/shared/ui/form-field'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/shared/ui/select'
import { useUploadImport } from '../api/use-upload-import'
import { validateImportFile } from '../model/validate-file'
import { FileDropzone } from './file-dropzone'

interface UploadImportFormProps {
  onUploaded: (batchId: string, isDuplicate: boolean) => void
}

export function UploadImportForm({ onUploaded }: UploadImportFormProps) {
  const sources = useDataSourceOptions()
  const upload = useUploadImport()
  const [dataSourceId, setDataSourceId] = useState<string | null>(null)
  const [file, setFile] = useState<File | null>(null)
  const [fileError, setFileError] = useState<string | null>(null)

  const sourceItems = (sources.data ?? []).map((s) => ({ value: s.id, label: s.name }))

  const selectFile = (selected: File) => {
    const error = validateImportFile(selected)
    setFileError(error)
    setFile(error ? null : selected)
  }

  const submit = (event: FormEvent) => {
    event.preventDefault()
    if (!dataSourceId || !file) return
    upload.mutate(
      { dataSourceId, file },
      { onSuccess: (accepted) => onUploaded(accepted.batchId, accepted.duplicate) },
    )
  }

  return (
    <form onSubmit={submit} className="flex flex-col gap-4">
      {sources.error && <ErrorBanner error={sources.error} />}
      {upload.error && <ErrorBanner error={upload.error} />}
      <FormField label="Источник данных">
        {(field) => (
          <Select
            items={sourceItems}
            value={dataSourceId}
            onValueChange={setDataSourceId}
            disabled={sources.isLoading}
          >
            <SelectTrigger {...field} className="w-full sm:w-72">
              <SelectValue placeholder="Выберите источник" />
            </SelectTrigger>
            <SelectContent>
              {sourceItems.map((item) => (
                <SelectItem key={item.value} value={item.value}>
                  {item.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        )}
      </FormField>
      <FileDropzone file={file} onFile={selectFile} />
      {fileError && (
        <Alert variant="destructive">
          <AlertDescription>{fileError}</AlertDescription>
        </Alert>
      )}
      <div>
        <Button type="submit" disabled={!dataSourceId || !file || upload.isPending}>
          Загрузить
        </Button>
      </div>
    </form>
  )
}
