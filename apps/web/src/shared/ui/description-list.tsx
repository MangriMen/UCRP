import type { ReactNode } from 'react'

interface DescriptionProps {
  label: string
  children: ReactNode
}

export function DescriptionRow({ label, children }: DescriptionProps) {
  return (
    <div className="flex flex-col gap-1 sm:flex-row sm:gap-4">
      <dt className="text-sm text-muted-foreground sm:w-48">{label}</dt>
      <dd className="text-sm">{children}</dd>
    </div>
  )
}

export function DescriptionField({ label, children }: DescriptionProps) {
  return (
    <div className="flex flex-col gap-0.5">
      <dt className="text-xs text-muted-foreground">{label}</dt>
      <dd className="text-sm break-words">{children}</dd>
    </div>
  )
}
