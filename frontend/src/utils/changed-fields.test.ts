import { describe, expect, it } from 'vitest'

import type { Competition } from '../types/competition'
import { changedFields } from './changed-fields'

const before: Competition = {
  id: 1,
  title: 'Кубок',
  sport: 'Футбол',
  location: 'Норильск',
  starts_on: '2026-12-01',
  capacity: 10,
  free_places: 10,
  description: '',
  created_at: '',
  updated_at: '',
}

// PATCH должен отправлять только изменённые поля, чтобы не перезаписать чужие правки
describe('changedFields', () => {
  it('без изменений — пустой объект', () => {
    expect(changedFields(before, { ...before })).toEqual({})
  })

  it('только изменённые поля', () => {
    expect(changedFields(before, { ...before, capacity: 24, description: 'Новое' })).toEqual({
      capacity: 24,
      description: 'Новое',
    })
  })
})
