import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { App } from '@/app/App'
import '@/app/index.css'

async function enableMocks() {
  if (import.meta.env.VITE_MOCKS !== '1') return
  const { worker } = await import('@/shared/api/mocks/browser')
  await worker.start({ onUnhandledFrame: 'bypass' })
}

void enableMocks().then(() => {
  createRoot(document.getElementById('root')!).render(
    <StrictMode>
      <App />
    </StrictMode>,
  )
})
