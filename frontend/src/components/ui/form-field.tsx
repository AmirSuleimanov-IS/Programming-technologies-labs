import type { ReactNode } from 'react'

interface FormFieldProps {
  label: string
  error?: string
  children: ReactNode
}

// Подпись, поле ввода и текст ошибки под ним
export function FormField({ label, error, children }: FormFieldProps) {
  return (
    <label className={error ? 'field field--invalid' : 'field'}>
      <span className="field__label">{label}</span>
      {children}
      {error && <span className="field__error">{error}</span>}
    </label>
  )
}
