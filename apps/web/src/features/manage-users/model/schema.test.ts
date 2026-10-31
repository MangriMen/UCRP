import { describe, expect, it } from 'vitest'
import { userFormSchema } from './schema'

const base = {
  email: 'user@ucrp.test',
  displayName: 'Иван',
  role: 'marketer',
  tenantId: '',
  isActive: true,
}

describe('userFormSchema', () => {
  it('requires a tenant only for the tenant role', () => {
    expect(userFormSchema.safeParse(base).success).toBe(true)
    expect(userFormSchema.safeParse({ ...base, role: 'tenant' }).success).toBe(false)
    expect(
      userFormSchema.safeParse({ ...base, role: 'tenant', tenantId: 'tenant-0001' }).success,
    ).toBe(true)
  })

  it('rejects an invalid email', () => {
    expect(userFormSchema.safeParse({ ...base, email: 'nope' }).success).toBe(false)
  })
})
