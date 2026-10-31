import { describe, expect, it } from 'vitest'
import { inviteFormSchema } from './schema'

const base = {
  email: 'user@ucrp.test',
  displayName: 'Иван',
  role: 'manager',
  tenantId: '',
}

describe('inviteFormSchema', () => {
  it('requires a tenant only for the tenant role', () => {
    expect(inviteFormSchema.safeParse(base).success).toBe(true)
    expect(inviteFormSchema.safeParse({ ...base, role: 'tenant' }).success).toBe(false)
    expect(
      inviteFormSchema.safeParse({ ...base, role: 'tenant', tenantId: 'tenant-0001' }).success,
    ).toBe(true)
  })

  it('accepts only roles the API can invite', () => {
    expect(inviteFormSchema.safeParse({ ...base, role: 'analyst' }).success).toBe(false)
  })

  it('rejects an invalid email', () => {
    expect(inviteFormSchema.safeParse({ ...base, email: 'nope' }).success).toBe(false)
  })
})
