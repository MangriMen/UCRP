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
import type { Role } from '@/entities/user'

export interface NavItem {
  title: string
  to: string
  icon: LucideIcon
  roles: Role[]
}

export interface NavGroup {
  label: string
  items: NavItem[]
}

const navigation: NavGroup[] = [
  {
    label: 'Аналитика',
    items: [
      {
        title: 'Дашборд',
        to: '/dashboard',
        icon: BarChart3,
        roles: ['manager', 'marketer', 'analyst'],
      },
      { title: 'Мой арендатор', to: '/tenant', icon: Store, roles: ['tenant'] },
      {
        title: 'Качество данных',
        to: '/data-quality',
        icon: ShieldCheck,
        roles: ['analyst', 'manager'],
      },
    ],
  },
  {
    label: 'Гости',
    items: [
      { title: 'Гости', to: '/guests', icon: UserSearch, roles: ['marketer', 'analyst'] },
      {
        title: 'Склейка профилей',
        to: '/identity/queue',
        icon: GitMerge,
        roles: ['analyst', 'marketer'],
      },
      { title: 'Запреты склейки', to: '/identity/blocks', icon: Ban, roles: ['analyst'] },
      {
        title: 'Сегменты',
        to: '/segments',
        icon: Layers,
        roles: ['marketer', 'analyst', 'manager'],
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
        roles: ['marketer', 'manager', 'analyst'],
      },
      {
        title: 'Кампании',
        to: '/campaigns',
        icon: Megaphone,
        roles: ['marketer', 'manager', 'analyst'],
      },
      {
        title: 'Лимиты контактов',
        to: '/settings/contact-limits',
        icon: Settings,
        roles: ['marketer', 'manager', 'analyst'],
      },
    ],
  },
  {
    label: 'Данные',
    items: [
      {
        title: 'Арендаторы и источники',
        to: '/data-sources',
        icon: Plug,
        roles: ['marketer', 'manager', 'analyst'],
      },
      { title: 'Импорт', to: '/import', icon: Upload, roles: ['analyst', 'manager'] },
    ],
  },
  {
    label: 'Администрирование',
    items: [
      { title: 'Пользователи', to: '/admin/users', icon: Users, roles: ['manager'] },
      {
        title: 'Справочники',
        to: '/admin/dictionaries',
        icon: Database,
        roles: ['analyst', 'marketer', 'manager'],
      },
      { title: 'Журнал аудита', to: '/admin/audit', icon: History, roles: ['manager', 'analyst'] },
    ],
  },
]

export function navigationFor(role: Role): NavGroup[] {
  return navigation
    .map((group) => ({ ...group, items: group.items.filter((item) => item.roles.includes(role)) }))
    .filter((group) => group.items.length > 0)
}
