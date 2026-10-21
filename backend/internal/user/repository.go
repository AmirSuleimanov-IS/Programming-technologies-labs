package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound   = errors.New("user not found")
	ErrEmailTaken = errors.New("email already registered")
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

const columns = `id, email, full_name, role, created_at, password_hash`

func scan(row pgx.Row) (User, error) {
	var u User
	err := row.Scan(&u.ID, &u.Email, &u.FullName, &u.Role, &u.CreatedAt, &u.PasswordHash)
	if errors.Is(err, pgx.ErrNoRows) {
		return u, ErrNotFound
	}
	return u, err
}

func (r *Repository) Create(ctx context.Context, u User) (User, error) {
	created, err := scan(r.db.QueryRow(ctx, `INSERT INTO users (email, full_name, password_hash, role)
		VALUES ($1, $2, $3, $4) RETURNING `+columns, u.Email, u.FullName, u.PasswordHash, u.Role))
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
		return User{}, ErrEmailTaken
	}
	return created, err
}

func (r *Repository) GetByEmail(ctx context.Context, email string) (User, error) {
	return scan(r.db.QueryRow(ctx, `SELECT `+columns+` FROM users WHERE email = $1`, email))
}

func (r *Repository) List(ctx context.Context) ([]User, error) {
	rows, err := r.db.Query(ctx, `SELECT `+columns+` FROM users ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()
	list := []User{}
	for rows.Next() {
		u, err := scan(rows)
		if err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		list = append(list, u)
	}
	return list, rows.Err()
}

func (r *Repository) GetByID(ctx context.Context, id int64) (User, error) {
	return scan(r.db.QueryRow(ctx, `SELECT `+columns+` FROM users WHERE id = $1`, id))
}
