export const MAX_FILE_BYTES = 10 * 1024 * 1024
export const ACCEPTED_EXTENSIONS = ['.csv', '.xlsx', '.xls']

export function validateImportFile(file: File): string | null {
  const name = file.name.toLowerCase()
  if (!ACCEPTED_EXTENSIONS.some((extension) => name.endsWith(extension))) {
    return 'Поддерживаются файлы CSV и Excel (.csv, .xlsx, .xls)'
  }
  if (file.size > MAX_FILE_BYTES) return 'Файл больше 10 МБ'
  return null
}
