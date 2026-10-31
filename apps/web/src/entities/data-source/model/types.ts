import type { Schema } from '@/shared/api/types'

export type DataSource = Schema<'DataSource'>
export type DataSourceLastImport = Schema<'DataSourceLastImport'>
export type DataSourceCreateRequest = Schema<'DataSourceCreateRequest'>
export type DataSourceUpdateRequest = Schema<'DataSourceUpdateRequest'>
export type DataSourceImportStatus = Schema<'ImportStatus'> | 'none'
export type DataSourceSort =
  | 'name'
  | '-name'
  | 'tenantName'
  | '-tenantName'
  | 'typeName'
  | '-typeName'
  | 'lastImportAt'
  | '-lastImportAt'
  | 'createdAt'
  | '-createdAt'
