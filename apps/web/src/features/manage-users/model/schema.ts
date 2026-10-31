import { z } from 'zod'

export const assignableRoles = ['marketer', 'manager', 'tenant', 'analyst'] as const

export const userFormSchema = z
  .object({
    email: z.email('Введите корректный email').max(320, 'Не более 320 символов'),
    displayName: z.string().trim().min(1, 'Введите имя').max(200, 'Не более 200 символов'),
    role: z.enum(assignableRoles),
    tenantId: z.string(),
    isActive: z.boolean(),
  })
  .refine((values) => values.role !== 'tenant' || values.tenantId !== '', {
    path: ['tenantId'],
    message: 'Выберите арендатора для роли «Арендатор»',
  })

export type UserFormValues = z.infer<typeof userFormSchema>
