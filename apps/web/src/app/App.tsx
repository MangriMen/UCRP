import { QueryClientProvider } from '@tanstack/react-query'
import { RouterProvider } from 'react-router'
import { endSession } from '@/entities/user'
import { setUnauthorizedHandler } from '@/shared/api/client'
import { queryClient } from '@/shared/api/query-client'
import { router } from './router'

setUnauthorizedHandler(() => endSession(queryClient))

export function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
    </QueryClientProvider>
  )
}
