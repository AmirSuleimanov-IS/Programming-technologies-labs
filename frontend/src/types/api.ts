// Общий формат ответов API (см. backend/README.md)

export interface FieldError {
  field: string
  message: string
}

export interface ApiErrorBody {
  code: string
  message: string
  details?: FieldError[]
}

export interface ApiSuccess<T, M = unknown> {
  data: T
  meta?: M
}
