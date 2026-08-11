package user

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStore struct{ pool *pgxpool.Pool }

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore { return &PostgresStore{pool: pool} }

const userColumns = `id, username, display_name, avatar_url, password_hash, role, active, created_at, updated_at`

func (s *PostgresStore) Create(ctx context.Context, item User) (User, error) {
	query := `INSERT INTO users (` + userColumns + `) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING ` + userColumns
	created, err := scanUser(s.pool.QueryRow(ctx, query, item.ID, item.Username, item.DisplayName, item.AvatarURL, item.PasswordHash, item.Role, item.Active, item.CreatedAt, item.UpdatedAt))
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
	query := `SELECT u.` + strings.ReplaceAll(userColumns, `, `, `, u.`) + ` FROM users u JOIN user_sessions s ON s.user_id = u.id WHERE s.token_hash = $1 AND s.expires_at > NOW() AND u.active = true`
	return scanUser(s.pool.QueryRow(ctx, query, tokenHash))
}

func (s *PostgresStore) DeleteSession(ctx context.Context, tokenHash string) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM user_sessions WHERE token_hash = $1`, tokenHash)
	return err
}

func (s *PostgresStore) ChangePassword(ctx context.Context, userID, passwordHash string, updatedAt time.Time, keepTokenHash string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin password change: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	result, err := tx.Exec(ctx, `UPDATE users SET password_hash=$2, updated_at=$3 WHERE id=$1`, userID, passwordHash, updatedAt)
	if err != nil {
		return fmt.Errorf("update password: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	if _, err := tx.Exec(ctx, `DELETE FROM user_sessions WHERE user_id=$1 AND token_hash<>$2`, userID, keepTokenHash); err != nil {
		return fmt.Errorf("revoke other sessions: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit password change: %w", err)
	}
	return nil
}

func (s *PostgresStore) UpdateProfile(ctx context.Context, item User) (User, error) {
	query := `UPDATE users SET username=$2, display_name=$3, updated_at=$4 WHERE id=$1 RETURNING ` + userColumns
	updated, err := scanUser(s.pool.QueryRow(ctx, query, item.ID, item.Username, item.DisplayName, item.UpdatedAt))
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return User{}, ErrUsernameExists
	}
	return updated, err
}

func (s *PostgresStore) UpdateAvatar(ctx context.Context, item User) (User, error) {
	return scanUser(s.pool.QueryRow(ctx, `UPDATE users SET avatar_url=$2, updated_at=$3 WHERE id=$1 RETURNING `+userColumns, item.ID, item.AvatarURL, item.UpdatedAt))
}

type scanner interface{ Scan(dest ...any) error }

func scanUser(row scanner) (User, error) {
	var item User
	if err := row.Scan(&item.ID, &item.Username, &item.DisplayName, &item.AvatarURL, &item.PasswordHash, &item.Role, &item.Active, &item.CreatedAt, &item.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, ErrNotFound
		}
		return User{}, fmt.Errorf("scan user: %w", err)
	}
	return item, nil
}
