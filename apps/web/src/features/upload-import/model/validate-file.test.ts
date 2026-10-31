import { describe, expect, it } from 'vitest'
import { MAX_FILE_BYTES, validateImportFile } from './validate-file'

const fileOf = (name: string, size = 10) => new File([new Uint8Array(size)], name)

describe('validateImportFile', () => {
  it('accepts csv and excel files regardless of extension case', () => {
    expect(validateImportFile(fileOf('a.csv'))).toBeNull()
    expect(validateImportFile(fileOf('A.XLSX'))).toBeNull()
    expect(validateImportFile(fileOf('a.xls'))).toBeNull()
  })

  it('rejects other formats', () => {
    expect(validateImportFile(fileOf('a.pdf'))).toMatch(/CSV и Excel/)
  })

  it('rejects files over the size limit', () => {
    expect(validateImportFile(fileOf('a.csv', MAX_FILE_BYTES + 1))).toMatch(/10 МБ/)
  })
})
