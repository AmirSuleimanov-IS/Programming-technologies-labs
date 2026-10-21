package competition

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound              = errors.New("competition not found")
	ErrCapacityBelowApproved = errors.New("capacity is below approved applications")
	ErrHasActiveApplications = errors.New("competition has active applications")
)

// Repository хранит соревнования в PostgreSQL.
type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// approvedCount — число одобренных заявок соревнования c.
const approvedCount = `(SELECT count(*) FROM applications a WHERE a.competition_id = c.id AND a.status = 'approved')`

const columns = `c.id, c.title, c.sport, c.location, to_char(c.starts_on, 'YYYY-MM-DD'), c.capacity, c.capacity - ` + approvedCount + `,
	c.description, c.created_at, c.updated_at`

func scan(row pgx.Row) (Competition, error) {
	var c Competition
	err := row.Scan(&c.ID, &c.Title, &c.Sport, &c.Location, &c.StartsOn, &c.Capacity, &c.FreePlaces, &c.Description, &c.CreatedAt, &c.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return c, ErrNotFound
	}
	return c, err
}

func (r *Repository) List(ctx context.Context, f Filter) ([]Competition, error) {
	rows, err := r.db.Query(ctx, `SELECT `+columns+` FROM competitions c
		WHERE ($1 = '' OR c.title ILIKE '%' || $1 || '%')
		  AND ($2 = '' OR lower(c.sport) = lower($2))
		ORDER BY c.starts_on, c.id`, f.Query, f.Sport)
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
	return scan(r.db.QueryRow(ctx, `SELECT `+columns+` FROM competitions c WHERE c.id = $1`, id))
}

func (r *Repository) Create(ctx context.Context, c Competition) (Competition, error) {
	return scan(r.db.QueryRow(ctx, `INSERT INTO competitions AS c (title, sport, location, starts_on, capacity, description)
		VALUES ($1, $2, $3, $4::date, $5, $6) RETURNING `+columns,
		c.Title, c.Sport, c.Location, c.StartsOn, c.Capacity, c.Description))
}

// Update сохраняет изменения. Лимит мест нельзя сделать меньше числа одобренных заявок:
// условие проверяется в том же UPDATE, поэтому одновременное одобрение его не обойдёт.
func (r *Repository) Update(ctx context.Context, c Competition) (Competition, error) {
	updated, err := scan(r.db.QueryRow(ctx, `UPDATE competitions AS c
		SET title = $2, sport = $3, location = $4, starts_on = $5::date, capacity = $6, description = $7, updated_at = now()
		WHERE c.id = $1 AND $6 >= `+approvedCount+` RETURNING `+columns,
		c.ID, c.Title, c.Sport, c.Location, c.StartsOn, c.Capacity, c.Description))
	if errors.Is(err, ErrNotFound) && r.exists(ctx, c.ID) {
		return Competition{}, ErrCapacityBelowApproved
	}
	return updated, err
}

// Delete удаляет соревнование, если у него нет активных заявок (на рассмотрении или одобренных).
// Отменённые и отклонённые заявки удаляются вместе с ним (ON DELETE CASCADE).
func (r *Repository) Delete(ctx context.Context, id int64) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM competitions WHERE id = $1 AND NOT EXISTS (
		SELECT 1 FROM applications WHERE competition_id = $1 AND status IN ('pending', 'approved'))`, id)
	if err != nil {
		return fmt.Errorf("delete competition: %w", err)
	}
	if tag.RowsAffected() == 0 {
		if r.exists(ctx, id) {
			return ErrHasActiveApplications
		}
		return ErrNotFound
	}
	return nil
}

func (r *Repository) exists(ctx context.Context, id int64) bool {
	var found bool
	_ = r.db.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM competitions WHERE id = $1)`, id).Scan(&found)
	return found
}
