package device

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStore struct {
	pool *pgxpool.Pool
}

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{pool: pool}
}

func (s *PostgresStore) Upsert(ctx context.Context, item Device) (Device, error) {
	const query = `
		INSERT INTO user_devices (
			id, user_id, platform, push_token, enabled, last_seen_at, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (user_id, push_token)
		DO UPDATE SET
			platform = EXCLUDED.platform,
			enabled = true,
			last_seen_at = EXCLUDED.last_seen_at,
			updated_at = EXCLUDED.updated_at
		RETURNING id, user_id, platform, push_token, enabled, last_seen_at, created_at, updated_at
	`

	updated, err := scanDevice(s.pool.QueryRow(ctx, query,
		item.ID,
		item.UserID,
		item.Platform,
		item.PushToken,
		item.Enabled,
		item.LastSeenAt,
		item.CreatedAt,
		item.UpdatedAt,
	))
	if err != nil {
		return Device{}, fmt.Errorf("upsert user device: %w", err)
	}

	return updated, nil
}

func (s *PostgresStore) ListEnabledByUser(ctx context.Context, userID string) ([]Device, error) {
	const query = `
		SELECT id, user_id, platform, push_token, enabled, last_seen_at, created_at, updated_at
		FROM user_devices
		WHERE user_id = $1 AND enabled = true
		ORDER BY updated_at DESC, id ASC
	`

	rows, err := s.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list enabled user devices: %w", err)
	}
	defer rows.Close()

	items := make([]Device, 0)
	for rows.Next() {
		item, err := scanDevice(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate enabled user devices: %w", err)
	}

	return items, nil
}

type deviceScanner interface {
	Scan(dest ...any) error
}

func scanDevice(scanner deviceScanner) (Device, error) {
	var item Device
	if err := scanner.Scan(
		&item.ID,
		&item.UserID,
		&item.Platform,
		&item.PushToken,
		&item.Enabled,
		&item.LastSeenAt,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		return Device{}, fmt.Errorf("scan user device: %w", err)
	}

	return item, nil
}
