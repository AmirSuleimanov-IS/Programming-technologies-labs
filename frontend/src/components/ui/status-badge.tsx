import { STATUS_TITLES } from '../../constants/application-status'
import type { ApplicationStatus } from '../../types/application'

export function StatusBadge({ status }: { status: ApplicationStatus }) {
  return <span className={`badge badge--${status}`}>{STATUS_TITLES[status]}</span>
}
