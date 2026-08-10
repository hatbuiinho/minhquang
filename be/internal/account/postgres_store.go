package account

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStore struct {
	pool *pgxpool.Pool
}

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{pool: pool}
}

func (s *PostgresStore) CreateUser(ctx context.Context, item User) (User, error) {
	const query = `
		INSERT INTO users (id, name, email, active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, name, email, active, created_at, updated_at
	`
	user, err := scanUser(s.pool.QueryRow(ctx, query, item.ID, item.Name, item.Email, item.Active, item.CreatedAt, item.UpdatedAt))
	if err != nil {
		return User{}, fmt.Errorf("create user: %w", err)
	}
	return user, nil
}

func (s *PostgresStore) ListUsers(ctx context.Context, activeOnly bool) ([]User, error) {
	query := `SELECT id, name, email, active, created_at, updated_at FROM users`
	args := []any{}
	if activeOnly {
		query += ` WHERE active = true`
	}
	query += ` ORDER BY name ASC, id ASC`
	rows, err := s.pool.Query(ctx, query, args...)
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
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate users: %w", err)
	}
	return items, nil
}

func (s *PostgresStore) CreateGroup(ctx context.Context, item Group) (Group, error) {
	const query = `
		INSERT INTO groups (id, name, description, active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, name, description, active, created_at, updated_at
	`
	group, err := scanGroup(s.pool.QueryRow(ctx, query, item.ID, item.Name, item.Description, item.Active, item.CreatedAt, item.UpdatedAt))
	if err != nil {
		return Group{}, fmt.Errorf("create group: %w", err)
	}
	return group, nil
}

func (s *PostgresStore) ListGroups(ctx context.Context, activeOnly bool) ([]Group, error) {
	query := `SELECT id, name, description, active, created_at, updated_at FROM groups`
	args := []any{}
	if activeOnly {
		query += ` WHERE active = true`
	}
	query += ` ORDER BY name ASC, id ASC`
	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list groups: %w", err)
	}
	defer rows.Close()

	items := make([]Group, 0)
	for rows.Next() {
		item, err := scanGroup(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate groups: %w", err)
	}
	return items, nil
}

func (s *PostgresStore) AddGroupMember(ctx context.Context, item GroupMember) error {
	const query = `
		INSERT INTO group_members (group_id, user_id, created_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (group_id, user_id) DO NOTHING
	`
	if _, err := s.pool.Exec(ctx, query, item.GroupID, item.UserID, item.CreatedAt); err != nil {
		return fmt.Errorf("add group member: %w", err)
	}
	return nil
}

func (s *PostgresStore) ListGroupMembers(ctx context.Context, groupIDs []string) ([]GroupMember, error) {
	if len(groupIDs) == 0 {
		return nil, nil
	}
	const query = `
		SELECT group_id, user_id, created_at
		FROM group_members
		WHERE group_id = ANY($1)
		ORDER BY group_id ASC, user_id ASC
	`
	rows, err := s.pool.Query(ctx, query, groupIDs)
	if err != nil {
		return nil, fmt.Errorf("list group members: %w", err)
	}
	defer rows.Close()

	items := make([]GroupMember, 0)
	for rows.Next() {
		var item GroupMember
		if err := rows.Scan(&item.GroupID, &item.UserID, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan group member: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate group members: %w", err)
	}
	return items, nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanUser(row scanner) (User, error) {
	var item User
	if err := row.Scan(&item.ID, &item.Name, &item.Email, &item.Active, &item.CreatedAt, &item.UpdatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return User{}, ErrNotFound
		}
		return User{}, fmt.Errorf("scan user: %w", err)
	}
	return item, nil
}

func scanGroup(row scanner) (Group, error) {
	var item Group
	if err := row.Scan(&item.ID, &item.Name, &item.Description, &item.Active, &item.CreatedAt, &item.UpdatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return Group{}, ErrNotFound
		}
		return Group{}, fmt.Errorf("scan group: %w", err)
	}
	return item, nil
}
