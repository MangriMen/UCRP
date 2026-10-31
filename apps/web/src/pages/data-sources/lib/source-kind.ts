const kinds = ['crm', 'pos', 'api'] as const

export function sourceKindLabel(typeCode: string | undefined, fallback: string | undefined) {
  const kind = kinds.find((k) => typeCode?.includes(k))
  return kind ? kind.toUpperCase() : (fallback ?? typeCode ?? '—')
}
