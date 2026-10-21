import type { Competition, CompetitionFilter, CompetitionInput } from '../types/competition'
import { request } from './http-client'

export async function fetchCompetitions(filter: CompetitionFilter): Promise<Competition[]> {
  const params = new URLSearchParams()
  if (filter.q) params.set('q', filter.q)
  if (filter.sport) params.set('sport', filter.sport)
  const query = params.toString()
  const res = await request<Competition[]>('GET', '/competitions' + (query ? `?${query}` : ''))
  return res.data
}

export async function createCompetition(input: CompetitionInput): Promise<Competition> {
  const res = await request<Competition>('POST', '/competitions', input)
  return res.data
}

// PATCH: отправляются только изменённые поля
export async function updateCompetition(
  id: number,
  changes: Partial<CompetitionInput>,
): Promise<Competition> {
  const res = await request<Competition>('PATCH', `/competitions/${id}`, changes)
  return res.data
}

export async function deleteCompetition(id: number): Promise<void> {
  await request<void>('DELETE', `/competitions/${id}`)
}
