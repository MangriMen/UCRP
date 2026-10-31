import { http, HttpResponse } from 'msw'
import { API, apiError, mockId, paged } from '@/shared/api/mocks/helpers'
import type { Project, ProjectRequest } from '../model/types'
import { mockProjects, PROJECT_WITH_EVENTS_ID } from './fixtures'

const projects: Project[] = structuredClone(mockProjects)

const isCodeTaken = (body: ProjectRequest, exceptId?: string) =>
  projects.some((p) => p.tenantId === body.tenantId && p.code === body.code && p.id !== exceptId)

export const projectHandlers = [
  http.get(`${API}/projects`, ({ request }) => {
    const tenantId = new URL(request.url).searchParams.get('tenant_id')
    const filtered = tenantId ? projects.filter((p) => p.tenantId === tenantId) : projects
    return HttpResponse.json(paged(request, filtered))
  }),

  http.post(`${API}/projects`, async ({ request }) => {
    const body = (await request.json()) as ProjectRequest
    if (isCodeTaken(body)) {
      return apiError(409, 'duplicate', 'Проект с таким кодом у арендатора уже существует')
    }
    const now = new Date().toISOString()
    const project: Project = {
      ...body,
      id: mockId('project-new'),
      tenantName: 'Арендатор',
      isActive: body.isActive ?? true,
      createdAt: now,
      updatedAt: now,
    }
    projects.push(project)
    return HttpResponse.json(project, { status: 201 })
  }),

  http.patch(`${API}/projects/:id`, async ({ params, request }) => {
    const project = projects.find((p) => p.id === params.id)
    if (!project) return apiError(404, 'not_found', 'Проект не найден')
    const body = (await request.json()) as ProjectRequest
    if (isCodeTaken(body, project.id)) {
      return apiError(409, 'duplicate', 'Проект с таким кодом у арендатора уже существует')
    }
    Object.assign(project, body, { updatedAt: new Date().toISOString() })
    return HttpResponse.json(project)
  }),

  http.delete(`${API}/projects/:id`, ({ params }) => {
    const index = projects.findIndex((p) => p.id === params.id)
    if (index === -1) return apiError(404, 'not_found', 'Проект не найден')
    if (params.id === PROJECT_WITH_EVENTS_ID) {
      return apiError(409, 'project_has_events', 'У проекта есть события, удаление невозможно')
    }
    projects.splice(index, 1)
    return new HttpResponse(null, { status: 204 })
  }),
]
