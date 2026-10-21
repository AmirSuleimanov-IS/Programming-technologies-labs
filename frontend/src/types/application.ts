export type ApplicationStatus = 'pending' | 'approved' | 'rejected' | 'cancelled'

export interface Application {
  id: number
  competition_id: number
  competition_title: string
  starts_on: string
  user_id: number
  user_full_name: string
  user_email: string
  status: ApplicationStatus
  created_at: string
  updated_at: string
}
