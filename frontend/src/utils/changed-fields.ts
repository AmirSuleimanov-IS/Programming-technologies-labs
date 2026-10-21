import type { Competition, CompetitionInput } from '../types/competition'

// Для PATCH оставляем только поля, которые пользователь действительно изменил
export function changedFields(
  before: Competition,
  after: CompetitionInput,
): Partial<CompetitionInput> {
  return Object.fromEntries(
    Object.entries(after).filter(([key, value]) => before[key as keyof CompetitionInput] !== value),
  )
}
