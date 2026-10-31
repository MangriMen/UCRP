import { createBrowserRouter, Outlet } from 'react-router'
import { HomePage } from '@/pages/home'
import { NotFoundPage } from '@/pages/not-found'

function Layout() {
  return (
    <div className="min-h-screen">
      <header className="border-b px-4 py-3 font-semibold">UCRP</header>
      <main className="p-4">
        <Outlet />
      </main>
    </div>
  )
}

export const router = createBrowserRouter([
  {
    element: <Layout />,
    children: [
      { index: true, element: <HomePage /> },
      { path: '*', element: <NotFoundPage /> },
    ],
  },
])
