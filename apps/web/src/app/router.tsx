import { createBrowserRouter, Link, Outlet } from 'react-router'

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

function HomePage() {
  return <h1 className="text-2xl font-semibold">UCRP</h1>
}

function NotFoundPage() {
  return (
    <div>
      <h1 className="text-2xl font-semibold">404</h1>
      <Link to="/" className="underline">
        На главную
      </Link>
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
