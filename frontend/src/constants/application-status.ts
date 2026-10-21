import type { ApplicationStatus } from '../types/application'

export const STATUS_TITLES: Record<ApplicationStatus, string> = {
  pending: 'На рассмотрении',
  approved: 'Одобрена',
  rejected: 'Отклонена',
  cancelled: 'Отменена',
}
