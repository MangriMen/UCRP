import { z } from 'zod'
import type { InvitableRole } from '@/entities/user'

export const invitableRoles = ['manager', 'tenant'] as const satisfies readonly InvitableRole[]

export const inviteFormSchema = z
  .object({
    email: z.email('Введите корректный email').max(320, 'Не более 320 символов'),
    displayName: z.string().trim().min(1, 'Введите имя').max(200, 'Не более 200 символов'),
    role: z.enum(invitableRoles),
    tenantId: z.string(),
  })
  .refine((values) => values.role !== 'tenant' || values.tenantId !== '', {
    path: ['tenantId'],
    message: 'Выберите арендатора для роли «Арендатор»',
  })

export type InviteFormValues = z.infer<typeof inviteFormSchema>
