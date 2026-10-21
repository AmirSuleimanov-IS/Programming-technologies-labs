export interface Competition {
  id: number
  title: string
  sport: string
  location: string
  starts_on: string // ГГГГ-ММ-ДД
  capacity: number
  description: string
  created_at: string
  updated_at: string
}

// Поля, которые пользователь вводит в форме
export type CompetitionInput = Pick<
  Competition,
  'title' | 'sport' | 'location' | 'starts_on' | 'capacity' | 'description'
>

export interface CompetitionFilter {
  q: string
  sport: string
}
