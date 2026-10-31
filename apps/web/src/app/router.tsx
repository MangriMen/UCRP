import { createBrowserRouter } from 'react-router'
import { RoleGuard } from '@/entities/user'
import { AdminDictionariesPage } from '@/pages/admin-dictionaries'
import { AdminUsersPage } from '@/pages/admin-users'
import { AuditLogPage } from '@/pages/audit-log'
import { CampaignDetailsPage } from '@/pages/campaign-details'
import { CampaignsPage } from '@/pages/campaigns'
import { ContactLimitsPage } from '@/pages/contact-limits'
import { DashboardPage } from '@/pages/dashboard'
import { DataQualityPage } from '@/pages/data-quality'
import { DataSourcesPage } from '@/pages/data-sources'
import { ExportsPage } from '@/pages/exports'
import { GuestProfilePage } from '@/pages/guest-profile'
import { GuestsPage } from '@/pages/guests'
import { ImportDetailsPage } from '@/pages/import-details'
import { ImportsPage } from '@/pages/imports'
import { MergeCandidatePage } from '@/pages/merge-candidate'
import { MergeCandidatesPage } from '@/pages/merge-candidates'
import { LoginPage } from '@/pages/login'
import { NotFoundPage } from '@/pages/not-found'
import { SegmentDetailsPage } from '@/pages/segment-details'
import { SegmentsPage } from '@/pages/segments'
import { TenantAggregatesPage } from '@/pages/tenant-aggregates'
import { MergeBlocksPage } from '@/pages/merge-blocks'
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
            element: <RoleGuard roles={['manager', 'marketer', 'analyst']} />,
            children: [{ path: 'dashboard', element: <DashboardPage /> }],
          },
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
          {
            element: <RoleGuard roles={['analyst', 'marketer']} />,
            children: [
              { path: 'identity/merge-candidates', element: <MergeCandidatesPage /> },
              { path: 'identity/merge-candidates/:id', element: <MergeCandidatePage /> },
            ],
          },
          {
            element: <RoleGuard roles={['marketer', 'analyst', 'manager']} />,
            children: [
              { path: 'segments', element: <SegmentsPage /> },
              { path: 'segments/:id', element: <SegmentDetailsPage /> },
              { path: 'exports', element: <ExportsPage /> },
              { path: 'campaigns', element: <CampaignsPage /> },
              { path: 'campaigns/:id', element: <CampaignDetailsPage /> },
              { path: 'settings/contact-limits', element: <ContactLimitsPage /> },
              { path: 'data-sources', element: <DataSourcesPage /> },
            ],
          },
          {
            element: <RoleGuard roles={['marketer', 'analyst', 'manager']} />,
            children: [{ path: 'admin/dictionaries', element: <AdminDictionariesPage /> }],
          },
          {
            element: <RoleGuard roles={['manager']} />,
            children: [{ path: 'admin/users', element: <AdminUsersPage /> }],
          },
          {
            element: <RoleGuard roles={['manager', 'analyst']} />,
            children: [{ path: 'admin/audit', element: <AuditLogPage /> }],
          },
          {
            element: <RoleGuard roles={['analyst', 'manager']} />,
            children: [{ path: 'data-quality', element: <DataQualityPage /> }],
          },
          {
            element: <RoleGuard roles={['tenant']} />,
            children: [{ path: 'tenant', element: <TenantAggregatesPage /> }],
          },
          {
            element: <RoleGuard roles={['analyst']} />,
            children: [{ path: 'identity/blocks', element: <MergeBlocksPage /> }],
          },
          { path: '*', element: <NotFoundPage /> },
        ],
      },
    ],
  },
])
