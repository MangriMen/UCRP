import { useCallback, useMemo } from 'react'
import { TENANT_ROLE_ID, type AccessSubject } from '@/entities/access'
import { useTenants } from '@/entities/tenant'
import { useAuth, useUsers } from '@/entities/user'
import type { DirectoryUser } from '../model/subject'

const DIRECTORY_PAGE_SIZE = 200
const SYSTEM_LABEL = 'Система'
const ALL_TENANTS_LABEL = 'Все арендаторы'

export interface SubjectDirectory {
  users: DirectoryUser[]
  tenants: { id: string; label: string }[]
  selfId: string | null
  userLabel: (id: string) => string
  subjectLabel: (subject: AccessSubject) => string
}

export function useSubjectDirectory(): SubjectDirectory {
  const { user: me, can } = useAuth()
  const users = useUsers({ pageSize: DIRECTORY_PAGE_SIZE }, { enabled: can('user.read') })
  const tenants = useTenants({ pageSize: DIRECTORY_PAGE_SIZE })

  const userList = useMemo(
    () =>
      (users.data?.items ?? [])
        .filter((user) => user.isActive && (user.role === 'tenant' || user.role === 'manager'))
        .map((user) => ({
          id: user.id,
          label: user.displayName || user.email,
          role: user.role,
          tenantId: user.tenantId ?? null,
        })),
    [users.data],
  )
  const tenantList = useMemo(
    () => (tenants.data?.items ?? []).map((tenant) => ({ id: tenant.id, label: tenant.name })),
    [tenants.data],
  )
  const allUsers = users.data?.items

  const userLabel = useCallback(
    (id: string) => {
      if (!id) return SYSTEM_LABEL
      const user = allUsers?.find((candidate) => candidate.id === id)
      return user ? user.displayName || user.email : id
    },
    [allUsers],
  )

  const subjectLabel = useCallback(
    ({ subjectType, subjectId }: AccessSubject) => {
      if (subjectType === 'role')
        return subjectId === TENANT_ROLE_ID ? ALL_TENANTS_LABEL : subjectId
      if (subjectType === 'user') return userLabel(subjectId)
      return tenantList.find((tenant) => tenant.id === subjectId)?.label ?? subjectId
    },
    [tenantList, userLabel],
  )

  return {
    users: userList,
    tenants: tenantList,
    selfId: me?.userId ?? null,
    userLabel,
    subjectLabel,
  }
}
