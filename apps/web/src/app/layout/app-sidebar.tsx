import { NavLink, useLocation } from 'react-router'
import { roleLabels, useAuth } from '@/entities/user'
import { LogoutButton } from '@/features/auth'
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from '@/shared/ui/sidebar'
import { navigationFor } from '../navigation'

export function AppSidebar() {
  const { user, role } = useAuth()
  const { pathname } = useLocation()
  const groups = role ? navigationFor(role) : []

  return (
    <Sidebar>
      <SidebarHeader className="px-4 py-3 text-lg font-semibold">UCRP</SidebarHeader>
      <SidebarContent>
        {groups.map((group) => (
          <SidebarGroup key={group.label}>
            <SidebarGroupLabel>{group.label}</SidebarGroupLabel>
            <SidebarGroupContent>
              <SidebarMenu>
                {group.items.map((item) => (
                  <SidebarMenuItem key={item.to}>
                    <SidebarMenuButton
                      render={<NavLink to={item.to} />}
                      isActive={pathname === item.to || pathname.startsWith(`${item.to}/`)}
                    >
                      <item.icon />
                      <span>{item.title}</span>
                    </SidebarMenuButton>
                  </SidebarMenuItem>
                ))}
              </SidebarMenu>
            </SidebarGroupContent>
          </SidebarGroup>
        ))}
      </SidebarContent>
      <SidebarFooter className="gap-2 p-3">
        {user && (
          <div className="min-w-0 text-sm">
            <p className="truncate font-medium">{user.displayName}</p>
            <p className="truncate text-xs text-muted-foreground">{roleLabels[user.role]}</p>
          </div>
        )}
        <LogoutButton />
      </SidebarFooter>
    </Sidebar>
  )
}
