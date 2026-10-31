import { useSegmentTemplates, type SegmentTemplate } from '@/entities/segment'
import { Button } from '@/shared/ui/button'
import { ErrorBanner } from '@/shared/ui/error-banner'
import { Skeleton } from '@/shared/ui/skeleton'

interface TemplateGalleryProps {
  selectedCode: string | null
  onSelect: (template: SegmentTemplate) => void
}

export function TemplateGallery({ selectedCode, onSelect }: TemplateGalleryProps) {
  const templates = useSegmentTemplates()

  if (templates.isLoading) return <Skeleton className="h-16 w-full" />
  if (templates.error) return <ErrorBanner error={templates.error} />

  return (
    <div className="flex flex-col gap-2">
      <p className="text-sm font-medium">Шаблоны</p>
      <div className="grid grid-cols-1 gap-2 sm:grid-cols-2">
        {templates.data?.map((template) => (
          <Button
            key={template.code}
            type="button"
            variant={template.code === selectedCode ? 'secondary' : 'outline'}
            className="h-auto flex-col items-start gap-0.5 py-2 text-left whitespace-normal"
            onClick={() => onSelect(template)}
          >
            <span>{template.name}</span>
            {template.description && (
              <span className="text-xs font-normal text-muted-foreground">
                {template.description}
              </span>
            )}
          </Button>
        ))}
      </div>
    </div>
  )
}
