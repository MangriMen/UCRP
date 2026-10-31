import type { StatusTone } from '@/shared/ui/status-badge'
import type { GrantTransition } from './lifecycle'
import type { GrantEffect, GrantScopeType, GrantStatus, GrantSubjectType } from './types'

export const grantStatusLabels: Record<GrantStatus, string> = {
  draft: 'Черновик',
  active: 'Действует',
  suspended: 'Приостановлен',
  revoked: 'Отозван',
  expired: 'Истёк',
}

export const grantStatusTones: Record<GrantStatus, StatusTone> = {
  draft: 'neutral',
  active: 'success',
  suspended: 'info',
  revoked: 'danger',
  expired: 'neutral',
}

export const grantTransitionLabels: Record<GrantTransition, string> = {
  activate: 'Активировать',
  suspend: 'Приостановить',
  revoke: 'Отозвать',
}

export const subjectTypeLabels: Record<GrantSubjectType, string> = {
  user: 'Пользователь',
  tenant: 'Арендатор',
  role: 'Роль',
}

export const effectLabels: Record<GrantEffect, string> = {
  allow: 'Разрешить',
  deny: 'Запретить',
}

export const scopeTypeLabels: Record<GrantScopeType, string> = {
  tenant: 'Арендатор',
  category: 'Категория',
  period_days: 'Глубина периода, дней',
  channel: 'Канал',
  k_threshold: 'Порог скрытия k',
  row_limit: 'Лимит строк',
  field_mask: 'Скрытое поле',
  metric: 'Показатель',
}

const presetLabels: Record<string, string> = {
  'tenant.none': 'Без доступа',
  'tenant.self_aggregates': 'Свои агрегаты',
  'tenant.self_aggregates_extended': 'Свои агрегаты, расширенный',
  'tenant.recommendations_inbound': 'Входящие рекомендации',
  'tenant.campaign_participant': 'Участник кампаний',
  'tenant.own_events': 'Свои события',
  'tenant.full_profile_read': 'Полный профиль гостя',
}

export const presetLabel = (code: string) => presetLabels[code] ?? code

const auditActionLabels: Record<string, string> = {
  'access.grant.create': 'Грант создан',
  'access.grant.update': 'Грант изменён',
  'access.grant.activate': 'Грант активирован',
  'access.grant.suspend': 'Грант приостановлен',
  'access.grant.revoke': 'Грант отозван',
  'access.simulate': 'Проверка доступа',
  'access.denied': 'Отказ в доступе',
}

export const auditActionLabel = (action: string) => auditActionLabels[action] ?? action
