import { describe, expect, it } from 'vitest'
import { sourceKindLabel } from './source-kind'

describe('sourceKindLabel', () => {
  it('maps known codes and falls back to the type name', () => {
    expect(sourceKindLabel('crm_api', 'CRM по API')).toBe('CRM')
    expect(sourceKindLabel('pos_api', undefined)).toBe('POS')
    expect(sourceKindLabel('api', undefined)).toBe('API')
    expect(sourceKindLabel('csv_excel', 'Файл CSV / Excel')).toBe('Файл CSV / Excel')
  })
})
