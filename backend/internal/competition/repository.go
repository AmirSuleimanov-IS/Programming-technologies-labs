package competition

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("competition not found")

// Repository хранит соревнования в PostgreSQL.
type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

const columns = `id, title, sport, location, to_char(starts_on, 'YYYY-MM-DD'), capacity, description, created_at, updated_at`

func scan(row pgx.Row) (Competition, error) {
	var c Competition
	err := row.Scan(&c.ID, &c.Title, &c.Sport, &c.Location, &c.StartsOn, &c.Capacity, &c.Description, &c.CreatedAt, &c.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return c, ErrNotFound
	}
	return c, err
}

func (r *Repository) List(ctx context.Context, f Filter) ([]Competition, error) {
	rows, err := r.db.Query(ctx, `SELECT `+columns+` FROM competitions
		WHERE ($1 = '' OR title ILIKE '%' || $1 || '%')
		  AND ($2 = '' OR lower(sport) = lower($2))
		ORDER BY starts_on, id`, f.Query, f.Sport)
	if err != nil {
		return nil, fmt.Errorf("list competitions: %w", err)
	}
	defer rows.Close()

	list := []Competition{}
	for rows.Next() {
		c, err := scan(rows)
		if err != nil {
			return nil, fmt.Errorf("scan competition: %w", err)
		}
		list = append(list, c)
	}
	return list, rows.Err()
}

func (r *Repository) Get(ctx context.Context, id int64) (Competition, error) {
	return scan(r.db.QueryRow(ctx, `SELECT `+columns+` FROM competitions WHERE id = $1`, id))
}

func (r *Repository) Create(ctx context.Context, c Competition) (Competition, error) {
	return scan(r.db.QueryRow(ctx, `INSERT INTO competitions (title, sport, location, starts_on, capacity, description)
		VALUES ($1, $2, $3, $4::date, $5, $6) RETURNING `+columns,
		c.Title, c.Sport, c.Location, c.StartsOn, c.Capacity, c.Description))
}

func (r *Repository) Update(ctx context.Context, c Competition) (Competition, error) {
	return scan(r.db.QueryRow(ctx, `UPDATE competitions
		SET title = $2, sport = $3, location = $4, starts_on = $5::date, capacity = $6, description = $7, updated_at = now()
		WHERE id = $1 RETURNING `+columns,
		c.ID, c.Title, c.Sport, c.Location, c.StartsOn, c.Capacity, c.Description))
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM competitions WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete competition: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
