import { createBrowserRouter } from 'react-router'
import { AccessGuard } from '@/entities/user'
import { AcceptInvitationPage } from '@/pages/accept-invitation'
import { AdminDictionariesPage } from '@/pages/admin-dictionaries'
import { AdminUsersPage } from '@/pages/admin-users'
import { AuditLogPage } from '@/pages/audit-log'
import { CampaignDetailsPage } from '@/pages/campaign-details'
import { CampaignsPage } from '@/pages/campaigns'
import { ContactLimitsPage } from '@/pages/contact-limits'
import { DashboardPage } from '@/pages/dashboard'
import { DataQualityPage } from '@/pages/data-quality'
import { DataSourcesPage } from '@/pages/data-sources'
import { ExportDetailsPage } from '@/pages/export-details'
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
import { TenantDetailsPage } from '@/pages/tenant-details'
import { TenantsPage } from '@/pages/tenants'
import { MergeBlocksPage } from '@/pages/merge-blocks'
import { AppLayout } from './layout/app-layout'
import { HomeRedirect } from './layout/home-redirect'
import { RouteError } from './layout/route-error'

export const router = createBrowserRouter([
  {
    errorElement: <RouteError />,
    children: [
      { path: '/login', element: <LoginPage /> },
      { path: '/invite/accept', element: <AcceptInvitationPage /> },
      {
        element: (
          <AccessGuard>
            <AppLayout />
          </AccessGuard>
        ),
        children: [
          { index: true, element: <HomeRedirect /> },
          {
            element: <AccessGuard permission="dashboard.read" />,
            children: [{ path: 'dashboard', element: <DashboardPage /> }],
          },
          {
            element: <AccessGuard permission="import.read" />,
            children: [
              { path: 'import', element: <ImportsPage /> },
              { path: 'import/:id', element: <ImportDetailsPage /> },
            ],
          },
          {
            element: <AccessGuard permission="guest.profile.search" />,
            children: [{ path: 'guests', element: <GuestsPage /> }],
          },
          {
            element: <AccessGuard permission="guest.profile.read" />,
            children: [{ path: 'guests/:id', element: <GuestProfilePage /> }],
          },
          {
            element: <AccessGuard permission="identity.queue.read" />,
            children: [
              { path: 'identity/merge-candidates', element: <MergeCandidatesPage /> },
              { path: 'identity/merge-candidates/:id', element: <MergeCandidatePage /> },
              { path: 'identity/merge-blocks', element: <MergeBlocksPage /> },
            ],
          },
          {
            element: <AccessGuard permission="segment.read" />,
            children: [
              { path: 'segments', element: <SegmentsPage /> },
              { path: 'segments/:id', element: <SegmentDetailsPage /> },
            ],
          },
          {
            element: <AccessGuard permission="export.read" />,
            children: [
              { path: 'exports', element: <ExportsPage /> },
              { path: 'exports/:id', element: <ExportDetailsPage /> },
            ],
          },
          {
            element: <AccessGuard permission="campaign.read" />,
            children: [
              { path: 'campaigns', element: <CampaignsPage /> },
              { path: 'campaigns/:id', element: <CampaignDetailsPage /> },
            ],
          },
          {
            element: <AccessGuard permission="contact_limit.read" />,
            children: [{ path: 'settings/contact-limits', element: <ContactLimitsPage /> }],
          },
          {
            element: <AccessGuard permission="data_source.read" />,
            children: [{ path: 'data-sources', element: <DataSourcesPage /> }],
          },
          {
            element: <AccessGuard permission="reference.read" />,
            children: [
              { path: 'tenants', element: <TenantsPage /> },
              { path: 'tenants/:id', element: <TenantDetailsPage /> },
              { path: 'admin/dictionaries', element: <AdminDictionariesPage /> },
            ],
          },
          {
            element: <AccessGuard permission="user.read" />,
            children: [{ path: 'admin/users', element: <AdminUsersPage /> }],
          },
          {
            element: <AccessGuard permission="audit.read" />,
            children: [{ path: 'admin/audit', element: <AuditLogPage /> }],
          },
          {
            element: <AccessGuard permission="data_quality.read" />,
            children: [{ path: 'data-quality', element: <DataQualityPage /> }],
          },
          {
            element: <AccessGuard permission="tenant.aggregates.own" tenantArea />,
            children: [{ path: 'tenant', element: <TenantAggregatesPage /> }],
          },
          { path: '*', element: <NotFoundPage /> },
        ],
      },
    ],
  },
])
