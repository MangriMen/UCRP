import { useState } from 'react'
import { useProjects, type Project } from '@/entities/project'
import { ActiveBadge } from '@/shared/ui/active-badge'
import { Button } from '@/shared/ui/button'
import { ConfirmDialog } from '@/shared/ui/confirm-dialog'
import { DataTable, type Column } from '@/shared/ui/data-table'
import { ErrorBanner } from '@/shared/ui/error-banner'
import { useDeleteProject } from '../api/use-project-mutations'
import { ProjectFormDialog } from './project-form-dialog'

interface ProjectsManagerProps {
  canEdit: boolean
}

interface FormState {
  project: Project | null
}

export function ProjectsManager({ canEdit }: ProjectsManagerProps) {
  const [page, setPage] = useState(1)
  const projects = useProjects({ page })
  const remove = useDeleteProject()
  const [form, setForm] = useState<FormState | null>(null)
  const [deleting, setDeleting] = useState<Project | null>(null)

  const columns: Column<Project>[] = [
    { id: 'code', header: 'Код', cell: (p) => p.code },
    { id: 'name', header: 'Название', cell: (p) => p.name },
    { id: 'tenant', header: 'Арендатор', cell: (p) => p.tenantName ?? '—' },
    { id: 'status', header: 'Статус', cell: (p) => <ActiveBadge isActive={p.isActive} /> },
    ...(canEdit
      ? [
          {
            id: 'actions',
            header: '',
            className: 'text-right',
            cell: (p: Project) => (
              <div className="flex justify-end gap-2">
                <Button variant="outline" size="sm" onClick={() => setForm({ project: p })}>
                  Изменить
                </Button>
                <Button variant="outline" size="sm" onClick={() => setDeleting(p)}>
                  Удалить
                </Button>
              </div>
            ),
          },
        ]
      : []),
  ]

  return (
    <div className="flex flex-col gap-4">
      {canEdit && (
        <div className="flex justify-end">
          <Button onClick={() => setForm({ project: null })}>Добавить проект</Button>
        </div>
      )}
      {remove.error && <ErrorBanner error={remove.error} />}
      <DataTable
        columns={columns}
        rows={projects.data?.items}
        getRowId={(p) => p.id}
        isLoading={projects.isLoading}
        error={projects.error}
        emptyTitle="Проектов пока нет"
        meta={projects.data?.meta}
        onPageChange={setPage}
      />
      {form && (
        <ProjectFormDialog
          open
          project={form.project}
          onOpenChange={(open) => !open && setForm(null)}
        />
      )}
      <ConfirmDialog
        open={deleting !== null}
        onOpenChange={(open) => !open && setDeleting(null)}
        title="Удалить проект?"
        description={`Проект «${deleting?.name ?? ''}» будет удалён. Если у него есть события, удаление будет отклонено.`}
        confirmLabel="Удалить"
        destructive
        isPending={remove.isPending}
        onConfirm={() => {
          if (deleting) remove.mutate(deleting.id, { onSettled: () => setDeleting(null) })
        }}
      />
    </div>
  )
}
