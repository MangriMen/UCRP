import type { ReactNode } from 'react'
import { cn } from 'cn'
import { Card, CardContent, CardDescription } from './card'
import { Skeleton } from './skeleton'

interface StatCardProps {
  label: string
  value: ReactNode
  hint?: string
  isLoading?: boolean
}

export function StatCard({ label, value, hint, isLoading }: StatCardProps) {
  return (
    <Card size="sm">
      <CardContent className="flex flex-col gap-1">
        <CardDescription>{label}</CardDescription>
        {isLoading ? (
          <Skeleton className="h-7 w-20" />
        ) : (
          <div className="text-2xl font-semibold tabular-nums">{value}</div>
        )}
        {hint && <p className="text-xs text-muted-foreground">{hint}</p>}
      </CardContent>
    </Card>
  )
}

export function StatGrid({ className, children }: { className?: string; children: ReactNode }) {
  return (
    <div className={cn('grid grid-cols-2 gap-3 md:grid-cols-3 xl:grid-cols-5', className)}>
      {children}
    </div>
  )
}
