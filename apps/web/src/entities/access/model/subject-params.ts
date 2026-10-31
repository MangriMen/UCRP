import type { AccessSubject, GrantSubjectType } from './types'

export const GRANT_SUBJECT_TYPES: GrantSubjectType[] = ['user', 'tenant', 'role']

export const parseSubjectType = (value: string | null) =>
  GRANT_SUBJECT_TYPES.find((type) => type === value)

export function subjectFromParams(params: URLSearchParams): AccessSubject | null {
  const subjectType = parseSubjectType(params.get('subject_type'))
  const subjectId = params.get('subject_id')
  return subjectType && subjectId ? { subjectType, subjectId } : null
}

export const subjectParams = ({ subjectType, subjectId }: AccessSubject) =>
  new URLSearchParams({ subject_type: subjectType, subject_id: subjectId })
