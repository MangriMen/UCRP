import { describe, expect, it } from 'vitest'
import { applyParamChanges, formatSort, nextSort, parsePageSize, parseSort } from './list-params'

describe('list params', () => {
  it('parses sort with direction prefix', () => {
    expect(parseSort('-events')).toEqual({ field: 'events', direction: 'desc' })
    expect(parseSort('name')).toEqual({ field: 'name', direction: 'asc' })
    expect(parseSort('')).toBeUndefined()
    expect(parseSort('-')).toBeUndefined()
  })

  it('formats sort back to the url value', () => {
    expect(formatSort({ field: 'events', direction: 'desc' })).toBe('-events')
    expect(formatSort({ field: 'name', direction: 'asc' })).toBe('name')
    expect(formatSort(undefined)).toBeUndefined()
  })

  it('cycles asc, desc, none for the same field and restarts for another', () => {
    const asc = nextSort(undefined, 'name')
    expect(asc).toEqual({ field: 'name', direction: 'asc' })
    const desc = nextSort(asc, 'name')
    expect(desc).toEqual({ field: 'name', direction: 'desc' })
    expect(nextSort(desc, 'name')).toBeUndefined()
    expect(nextSort(desc, 'events')).toEqual({ field: 'events', direction: 'asc' })
  })

  it('accepts only known page sizes', () => {
    expect(parsePageSize('100')).toBe(100)
    expect(parsePageSize('7')).toBe(50)
    expect(parsePageSize(null)).toBe(50)
  })

  it('resets the page when filters change but keeps an explicit page', () => {
    const current = new URLSearchParams('page=3&q=a')
    expect(applyParamChanges(current, { q: 'b' }).toString()).toBe('q=b')
    expect(applyParamChanges(current, { page: '4' }, false).toString()).toBe('page=4&q=a')
    expect(applyParamChanges(current, { q: null }).toString()).toBe('')
  })
})
