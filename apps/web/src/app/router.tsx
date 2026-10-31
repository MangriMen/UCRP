import { createBrowserRouter } from 'react-router'
import { RoleGuard } from '@/entities/user'
import { GuestProfilePage } from '@/pages/guest-profile'
import { GuestsPage } from '@/pages/guests'
import { ImportDetailsPage } from '@/pages/import-details'
import { ImportsPage } from '@/pages/imports'
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
          {
            element: <RoleGuard roles={['analyst', 'manager']} />,
            children: [
              { path: 'import', element: <ImportsPage /> },
              { path: 'import/:id', element: <ImportDetailsPage /> },
            ],
          },
          {
            element: <RoleGuard roles={['marketer', 'analyst']} />,
            children: [
              { path: 'guests', element: <GuestsPage /> },
              { path: 'guests/:id', element: <GuestProfilePage /> },
            ],
          },
          { path: '*', element: <NotFoundPage /> },
        ],
      },
    ],
  },
])
