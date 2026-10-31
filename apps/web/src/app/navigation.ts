import {
  Ban,
  BarChart3,
  Database,
  FileDown,
  GitMerge,
  History,
  Layers,
  Megaphone,
  Plug,
  Settings,
  ShieldCheck,
  Store,
  Upload,
  Users,
  UserSearch,
  type LucideIcon,
} from 'lucide-react'
import type { Me, PermissionCode } from '@/entities/user'

export interface NavItem {
  title: string
  to: string
  icon: LucideIcon
  permission: PermissionCode
  tenantArea?: boolean
}

export interface NavGroup {
  label: string
  items: NavItem[]
}

const navigation: NavGroup[] = [
  {
    label: 'Обзор',
    items: [
      {
        title: 'Дашборд',
        to: '/dashboard',
        icon: BarChart3,
        permission: 'dashboard.read',
      },
      {
        title: 'Мой арендатор',
        to: '/tenant',
        icon: Store,
        permission: 'tenant.aggregates.own',
        tenantArea: true,
      },
      {
        title: 'Качество данных',
        to: '/data-quality',
        icon: ShieldCheck,
        permission: 'data_quality.read',
      },
    ],
  },
  {
    label: 'Данные',
    items: [
      { title: 'Арендаторы', to: '/tenants', icon: Store, permission: 'reference.read' },
      {
        title: 'Источники данных',
        to: '/data-sources',
        icon: Plug,
        permission: 'data_source.read',
      },
      { title: 'Импорт', to: '/import', icon: Upload, permission: 'import.read' },
    ],
  },
  {
    label: 'Гости',
    items: [
      { title: 'Гости', to: '/guests', icon: UserSearch, permission: 'guest.profile.search' },
      {
        title: 'Сегменты',
        to: '/segments',
        icon: Layers,
        permission: 'segment.read',
      },
      {
        title: 'Объединение профилей',
        to: '/identity/merge-candidates',
        icon: GitMerge,
        permission: 'identity.queue.read',
      },
    ],
  },
  {
    label: 'Коммуникации',
    items: [
      {
        title: 'Выгрузки',
        to: '/exports',
        icon: FileDown,
        permission: 'export.read',
      },
      {
        title: 'Кампании',
        to: '/campaigns',
        icon: Megaphone,
        permission: 'campaign.read',
      },
      {
        title: 'Лимиты контактов',
        to: '/settings/contact-limits',
        icon: Settings,
        permission: 'contact_limit.read',
      },
    ],
  },
  {
    label: 'Администрирование',
    items: [
      { title: 'Пользователи', to: '/admin/users', icon: Users, permission: 'user.read' },
      {
        title: 'Справочники',
        to: '/admin/dictionaries',
        icon: Database,
        permission: 'reference.read',
      },
      {
        title: 'Запреты объединения',
        to: '/identity/merge-blocks',
        icon: Ban,
        permission: 'identity.queue.read',
      },
      { title: 'Журнал аудита', to: '/admin/audit', icon: History, permission: 'audit.read' },
    ],
  },
]

export function navigationFor(user: Me): NavGroup[] {
  const isTenant = user.role === 'tenant'
  const visible = (item: NavItem) =>
    user.permissions.includes(item.permission) && (item.tenantArea ?? false) === isTenant
  return navigation
    .map((group) => ({ ...group, items: group.items.filter(visible) }))
    .filter((group) => group.items.length > 0)
}
