interface CompetitionFiltersProps {
  query: string
  sport: string
  onQueryChange: (value: string) => void
  onSportChange: (value: string) => void
}

export function CompetitionFilters({
  query,
  sport,
  onQueryChange,
  onSportChange,
}: CompetitionFiltersProps) {
  return (
    <div className="filters">
      <input
        placeholder="Поиск по названию"
        value={query}
        onChange={(e) => onQueryChange(e.target.value)}
      />
      <input
        placeholder="Вид спорта"
        value={sport}
        onChange={(e) => onSportChange(e.target.value)}
      />
    </div>
  )
}
