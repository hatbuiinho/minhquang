package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStore struct{ pool *pgxpool.Pool }

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore { return &PostgresStore{pool: pool} }

const userColumns = `id, username, display_name, password_hash, role, active, created_at, updated_at`

func (s *PostgresStore) Create(ctx context.Context, item User) (User, error) {
	query := `INSERT INTO users (` + userColumns + `) VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING ` + userColumns
	created, err := scanUser(s.pool.QueryRow(ctx, query, item.ID, item.Username, item.DisplayName, item.PasswordHash, item.Role, item.Active, item.CreatedAt, item.UpdatedAt))
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return User{}, ErrUsernameExists
	}
	return created, err
}

func (s *PostgresStore) List(ctx context.Context) ([]User, error) {
	rows, err := s.pool.Query(ctx, `SELECT `+userColumns+` FROM users ORDER BY display_name ASC`)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()
	items := make([]User, 0)
	for rows.Next() {
		item, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *PostgresStore) FindByUsername(ctx context.Context, username string) (User, error) {
	return scanUser(s.pool.QueryRow(ctx, `SELECT `+userColumns+` FROM users WHERE username = $1`, username))
}

func (s *PostgresStore) CreateSession(ctx context.Context, session Session) error {
	_, err := s.pool.Exec(ctx, `INSERT INTO user_sessions (token_hash, user_id, expires_at, created_at) VALUES ($1,$2,$3,$4)`, session.TokenHash, session.UserID, session.ExpiresAt, session.CreatedAt)
	return err
}

func (s *PostgresStore) UserBySession(ctx context.Context, tokenHash string) (User, error) {
	query := `SELECT u.id, u.username, u.display_name, u.password_hash, u.role, u.active, u.created_at, u.updated_at FROM users u JOIN user_sessions s ON s.user_id = u.id WHERE s.token_hash = $1 AND s.expires_at > NOW() AND u.active = true`
	return scanUser(s.pool.QueryRow(ctx, query, tokenHash))
}

func (s *PostgresStore) DeleteSession(ctx context.Context, tokenHash string) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM user_sessions WHERE token_hash = $1`, tokenHash)
	return err
}

type scanner interface{ Scan(dest ...any) error }

func scanUser(row scanner) (User, error) {
	var item User
	if err := row.Scan(&item.ID, &item.Username, &item.DisplayName, &item.PasswordHash, &item.Role, &item.Active, &item.CreatedAt, &item.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, ErrNotFound
		}
		return User{}, fmt.Errorf("scan user: %w", err)
	}
	return item, nil
}
