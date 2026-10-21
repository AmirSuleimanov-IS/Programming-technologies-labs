import { useState, type FormEvent } from 'react'

import { ApiError } from '../../api/http-client'
import { SPORT_SUGGESTIONS } from '../../constants/sports'
import type { Competition, CompetitionInput } from '../../types/competition'
import { todayIso } from '../../utils/format-date'
import { FormField } from '../ui/form-field'

interface CompetitionFormProps {
  initial?: Competition
  onSubmit: (input: CompetitionInput) => Promise<void>
  onCancel: () => void
}

const EMPTY: CompetitionInput = {
  title: '',
  sport: '',
  location: '',
  starts_on: '',
  capacity: 10,
  description: '',
}

// Форма создания и изменения соревнования. Ошибки полей приходят с бэкенда (422).
export function CompetitionForm({ initial, onSubmit, onCancel }: CompetitionFormProps) {
  const [values, setValues] = useState<CompetitionInput>(initial ?? EMPTY)
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({})
  const [formError, setFormError] = useState<string | null>(null)
  const [saving, setSaving] = useState(false)

  const set =
    <K extends keyof CompetitionInput>(key: K) =>
    (value: CompetitionInput[K]) =>
      setValues((v) => ({ ...v, [key]: value }))

  async function handleSubmit(event: FormEvent) {
    event.preventDefault()
    setSaving(true)
    setFormError(null)
    setFieldErrors({})
    try {
      await onSubmit(values)
    } catch (err) {
      if (err instanceof ApiError && err.details.length > 0) {
        setFieldErrors(Object.fromEntries(err.details.map((d) => [d.field, d.message])))
      } else {
        setFormError(err instanceof Error ? err.message : 'Не удалось сохранить')
      }
    } finally {
      setSaving(false)
    }
  }

  return (
    <form className="panel form" onSubmit={handleSubmit} noValidate>
      <h2>{initial ? 'Изменить соревнование' : 'Новое соревнование'}</h2>
      {formError && <p className="alert alert--error">{formError}</p>}
      <FormField label="Название" error={fieldErrors.title}>
        <input value={values.title} onChange={(e) => set('title')(e.target.value)} />
      </FormField>
      <div className="form__row">
        <FormField label="Вид спорта" error={fieldErrors.sport}>
          <input
            list="sport-suggestions"
            value={values.sport}
            onChange={(e) => set('sport')(e.target.value)}
          />
          <datalist id="sport-suggestions">
            {SPORT_SUGGESTIONS.map((s) => (
              <option key={s} value={s} />
            ))}
          </datalist>
        </FormField>
        <FormField label="Дата начала" error={fieldErrors.starts_on}>
          <input
            type="date"
            min={todayIso()}
            value={values.starts_on}
            onChange={(e) => set('starts_on')(e.target.value)}
          />
        </FormField>
        <FormField label="Мест" error={fieldErrors.capacity}>
          <input
            type="number"
            min={1}
            max={1000}
            value={values.capacity}
            onChange={(e) => set('capacity')(Number(e.target.value))}
          />
        </FormField>
      </div>
      <FormField label="Место проведения" error={fieldErrors.location}>
        <input value={values.location} onChange={(e) => set('location')(e.target.value)} />
      </FormField>
      <FormField label="Описание" error={fieldErrors.description}>
        <textarea
          rows={3}
          value={values.description}
          onChange={(e) => set('description')(e.target.value)}
        />
      </FormField>
      <div className="form__actions">
        <button type="button" className="button button--ghost" onClick={onCancel}>
          Отмена
        </button>
        <button type="submit" className="button" disabled={saving}>
          {saving ? 'Сохранение…' : 'Сохранить'}
        </button>
      </div>
    </form>
  )
}
