package application

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound            = errors.New("application not found")
	ErrCompetitionNotFound = errors.New("competition not found")
	ErrDuplicate           = errors.New("active application already exists")
)

// Repository хранит заявки в PostgreSQL.
type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

const selectApplications = `SELECT a.id, a.competition_id, c.title, to_char(c.starts_on, 'YYYY-MM-DD'),
		a.user_id, u.full_name, u.email, a.status, a.created_at, a.updated_at
	FROM applications a
	JOIN competitions c ON c.id = a.competition_id
	JOIN users u ON u.id = a.user_id`

func scanApplication(row pgx.Row) (Application, error) {
	var a Application
	err := row.Scan(&a.ID, &a.CompetitionID, &a.CompetitionTitle, &a.StartsOn,
		&a.UserID, &a.UserFullName, &a.UserEmail, &a.Status, &a.CreatedAt, &a.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return a, ErrNotFound
	}
	return a, err
}

func collect(rows pgx.Rows, err error) ([]Application, error) {
	if err != nil {
		return nil, fmt.Errorf("list applications: %w", err)
	}
	defer rows.Close()
	list := []Application{}
	for rows.Next() {
		a, err := scanApplication(rows)
		if err != nil {
			return nil, fmt.Errorf("scan application: %w", err)
		}
		list = append(list, a)
	}
	return list, rows.Err()
}

func (r *Repository) ListByUser(ctx context.Context, userID int64) ([]Application, error) {
	return collect(r.db.Query(ctx, selectApplications+` WHERE a.user_id = $1 ORDER BY c.starts_on, a.id`, userID))
}

func (r *Repository) List(ctx context.Context, f Filter) ([]Application, error) {
	return collect(r.db.Query(ctx, selectApplications+`
		WHERE ($1 = '' OR a.status = $1) AND ($2 = 0 OR a.competition_id = $2)
		ORDER BY a.created_at DESC, a.id DESC`, string(f.Status), f.CompetitionID))
}

// InTx выполняет fn в транзакции: при ошибке изменения откатываются.
func (r *Repository) InTx(ctx context.Context, fn func(tx Tx) error) error {
	return pgx.BeginFunc(ctx, r.db, func(t pgx.Tx) error {
		return fn(&pgTx{tx: t})
	})
}

// Tx — операции, которые сервис выполняет внутри одной транзакции.
type Tx interface {
	LockSlot(ctx context.Context, competitionID int64) (Slot, error)
	GetForUpdate(ctx context.Context, id int64) (Application, error)
	Create(ctx context.Context, userID, competitionID int64) (Application, error)
	SetStatus(ctx context.Context, id int64, status Status) (Application, error)
}

type pgTx struct {
	tx pgx.Tx
}

// LockSlot блокирует строку соревнования до конца транзакции (FOR UPDATE), поэтому
// два одновременных одобрения не займут одно и то же последнее место.
func (p *pgTx) LockSlot(ctx context.Context, competitionID int64) (Slot, error) {
	s := Slot{CompetitionID: competitionID}
	err := p.tx.QueryRow(ctx, `SELECT starts_on, capacity FROM competitions WHERE id = $1 FOR UPDATE`, competitionID).
		Scan(&s.StartsOn, &s.Capacity)
	if errors.Is(err, pgx.ErrNoRows) {
		return s, ErrCompetitionNotFound
	}
	if err != nil {
		return s, err
	}
	err = p.tx.QueryRow(ctx, `SELECT count(*) FROM applications WHERE competition_id = $1 AND status = 'approved'`,
		competitionID).Scan(&s.Approved)
	return s, err
}

func (p *pgTx) GetForUpdate(ctx context.Context, id int64) (Application, error) {
	var lockedID int64
	err := p.tx.QueryRow(ctx, `SELECT id FROM applications WHERE id = $1 FOR UPDATE`, id).Scan(&lockedID)
	if errors.Is(err, pgx.ErrNoRows) {
		return Application{}, ErrNotFound
	}
	if err != nil {
		return Application{}, err
	}
	return scanApplication(p.tx.QueryRow(ctx, selectApplications+` WHERE a.id = $1`, id))
}

func (p *pgTx) Create(ctx context.Context, userID, competitionID int64) (Application, error) {
	var id int64
	err := p.tx.QueryRow(ctx, `INSERT INTO applications (competition_id, user_id, status)
		VALUES ($1, $2, 'pending') RETURNING id`, competitionID, userID).Scan(&id)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" { // уникальный индекс applications_one_active_idx
		return Application{}, ErrDuplicate
	}
	if err != nil {
		return Application{}, err
	}
	return scanApplication(p.tx.QueryRow(ctx, selectApplications+` WHERE a.id = $1`, id))
}

func (p *pgTx) SetStatus(ctx context.Context, id int64, status Status) (Application, error) {
	if _, err := p.tx.Exec(ctx, `UPDATE applications SET status = $2, updated_at = now() WHERE id = $1`, id, status); err != nil {
		return Application{}, fmt.Errorf("update application status: %w", err)
	}
	return scanApplication(p.tx.QueryRow(ctx, selectApplications+` WHERE a.id = $1`, id))
}
