import type { Schema } from '@/shared/api/types'

export type ImportBatch = Schema<'ImportBatch'>
export type ImportStatus = Schema<'ImportStatus'>
export type RejectedRow = Schema<'RejectedRow'>
export type RejectedRowCode = RejectedRow['errorCode']
