import type { ApiErrorBody, ApiSuccess, FieldError } from '../types/api'

const API_URL = import.meta.env.VITE_API_URL ?? '/api'

// Ошибка API в общем формате: код ответа, машинный код, текст и ошибки полей
export class ApiError extends Error {
  readonly status: number
  readonly code: string
  readonly details: FieldError[]

  constructor(status: number, body: ApiErrorBody) {
    super(body.message)
    this.status = status
    this.code = body.code
    this.details = body.details ?? []
  }
}

type Method = 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE'

export async function request<T, M = unknown>(
  method: Method,
  path: string,
  body?: unknown,
): Promise<ApiSuccess<T, M>> {
  let response: Response
  try {
    response = await fetch(API_URL + path, {
      method,
      headers: body === undefined ? undefined : { 'Content-Type': 'application/json' },
      body: body === undefined ? undefined : JSON.stringify(body),
    })
  } catch {
    throw new ApiError(0, { code: 'NETWORK_ERROR', message: 'Сервер недоступен' })
  }

  if (response.status === 204) {
    return { data: undefined as T }
  }
  const json = await response.json().catch(() => null)
  if (!response.ok) {
    throw new ApiError(
      response.status,
      json?.error ?? { code: 'HTTP_ERROR', message: `Ошибка ${response.status}` },
    )
  }
  return json as ApiSuccess<T, M>
}
