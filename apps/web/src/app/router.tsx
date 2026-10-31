import { createBrowserRouter, Outlet } from 'react-router'
import { RoleGuard } from '@/entities/user'
import { LogoutButton } from '@/features/auth'
import { HomePage } from '@/pages/home'
import { LoginPage } from '@/pages/login'
import { NotFoundPage } from '@/pages/not-found'

function Layout() {
  return (
    <div className="min-h-screen">
      <header className="flex items-center justify-between border-b px-4 py-3">
        <span className="font-semibold">UCRP</span>
        <LogoutButton />
      </header>
      <main className="p-4">
        <Outlet />
      </main>
    </div>
  )
}

export const router = createBrowserRouter([
  { path: '/login', element: <LoginPage /> },
  {
    element: (
      <RoleGuard>
        <Layout />
      </RoleGuard>
    ),
    children: [
      { index: true, element: <HomePage /> },
      { path: '*', element: <NotFoundPage /> },
    ],
  },
])
