import { z } from 'zod'
import type { Role } from '@/entities/user'

export const assignableRoles = ['marketer', 'manager', 'tenant', 'analyst'] as const

export const recoverableRoles: readonly Role[] = ['manager', 'tenant']

export const renameUserSchema = z.object({
  displayName: z.string().trim().min(1, 'Введите имя').max(200, 'Не более 200 символов'),
})

export type RenameUserValues = z.infer<typeof renameUserSchema>
