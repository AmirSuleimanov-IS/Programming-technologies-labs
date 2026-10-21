import type { Application, ApplicationStatus } from '../types/application'
import { request } from './http-client'

export async function fetchMyApplications(): Promise<Application[]> {
  const res = await request<Application[]>('GET', '/applications/my')
  return res.data
}

// Все заявки (только ADMIN), можно отфильтровать по статусу
export async function fetchApplications(status = ''): Promise<Application[]> {
  const res = await request<Application[]>(
    'GET',
    '/applications' + (status ? `?status=${status}` : ''),
  )
  return res.data
}

export async function submitApplication(competitionId: number): Promise<Application> {
  const res = await request<Application>('POST', '/applications', { competition_id: competitionId })
  return res.data
}

export async function changeApplicationStatus(
  id: number,
  status: Exclude<ApplicationStatus, 'pending'>,
): Promise<Application> {
  const res = await request<Application>('PATCH', `/applications/${id}`, { status })
  return res.data
}
