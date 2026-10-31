import { z } from 'zod'

export const acceptInvitationSchema = z
  .object({
    password: z.string().min(8, 'Не менее 8 символов').max(72, 'Не более 72 символов'),
    confirmation: z.string(),
  })
  .refine((values) => values.password === values.confirmation, {
    path: ['confirmation'],
    message: 'Пароли не совпадают',
  })

export type AcceptInvitationValues = z.infer<typeof acceptInvitationSchema>
