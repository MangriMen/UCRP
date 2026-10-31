import { Link } from 'react-router'

export function NotFoundPage() {
  return (
    <div>
      <h1 className="text-2xl font-semibold">404</h1>
      <Link to="/" className="underline">
        На главную
      </Link>
    </div>
  )
}
