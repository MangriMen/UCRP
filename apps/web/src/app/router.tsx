import { createBrowserRouter } from 'react-router'
import { RoleGuard } from '@/entities/user'
import { LoginPage } from '@/pages/login'
import { NotFoundPage } from '@/pages/not-found'
import { AppLayout } from './layout/app-layout'
import { HomeRedirect } from './layout/home-redirect'
import { RouteError } from './layout/route-error'

export const router = createBrowserRouter([
  {
    errorElement: <RouteError />,
    children: [
      { path: '/login', element: <LoginPage /> },
      {
        element: (
          <RoleGuard>
            <AppLayout />
          </RoleGuard>
        ),
        children: [
          { index: true, element: <HomeRedirect /> },
          { path: '*', element: <NotFoundPage /> },
        ],
      },
    ],
  },
])
