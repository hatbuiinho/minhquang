package department

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

const departmentColumns = `d.id, d.name, d.active, COUNT(v.id), COUNT(v.id) FILTER (WHERE v.departure_date IS NULL OR v.departure_date >= (CURRENT_TIMESTAMP AT TIME ZONE 'Asia/Ho_Chi_Minh')::date), d.created_at, d.updated_at`

func (s *PostgresStore) Create(ctx context.Context, item Department, key string) (Department, error) {
	_, err := s.pool.Exec(ctx, `INSERT INTO departments (id, name, search_key, active, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6)`, item.ID, item.Name, key, item.Active, item.CreatedAt, item.UpdatedAt)
	if mapped := mapWriteError(err); mapped != nil {
		return Department{}, mapped
	}
	return item, nil
}

func (s *PostgresStore) List(ctx context.Context, options ListOptions) ([]Department, error) {
	const query = `
		SELECT ` + departmentColumns + `
		FROM departments d
		LEFT JOIN volunteers v ON v.department_id = d.id
		WHERE ($1 = '' OR unaccent(lower(d.name)) LIKE '%' || unaccent(lower($1)) || '%')
		  AND ($2::boolean IS NULL OR d.active = $2)
		GROUP BY d.id
		ORDER BY d.active DESC, 5 DESC, d.name ASC
		LIMIT NULLIF($3, 0)
	`
	rows, err := s.pool.Query(ctx, query, options.Query, options.Active, options.Limit)
	if err != nil {
		return nil, fmt.Errorf("list departments: %w", err)
	}
	defer rows.Close()
	items := make([]Department, 0)
	for rows.Next() {
		item, err := scanDepartment(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *PostgresStore) Get(ctx context.Context, id string) (Department, error) {
	return scanDepartment(s.pool.QueryRow(ctx, `SELECT `+departmentColumns+` FROM departments d LEFT JOIN volunteers v ON v.department_id=d.id WHERE d.id=$1 GROUP BY d.id`, id))
}

func (s *PostgresStore) FindBySearchKey(ctx context.Context, key string) (Department, error) {
	return scanDepartment(s.pool.QueryRow(ctx, `SELECT `+departmentColumns+` FROM departments d LEFT JOIN volunteers v ON v.department_id=d.id WHERE d.search_key=$1 GROUP BY d.id`, key))
}

func (s *PostgresStore) Update(ctx context.Context, item Department, key string) (Department, error) {
	result, err := s.pool.Exec(ctx, `UPDATE departments SET name=$2, search_key=$3, updated_at=$4 WHERE id=$1`, item.ID, item.Name, key, item.UpdatedAt)
	if mapped := mapWriteError(err); mapped != nil {
		return Department{}, mapped
	}
	if result.RowsAffected() == 0 {
		return Department{}, ErrNotFound
	}
	return s.Get(ctx, item.ID)
}

func (s *PostgresStore) SetActive(ctx context.Context, id string, active bool) (Department, error) {
	result, err := s.pool.Exec(ctx, `UPDATE departments SET active=$2, updated_at=NOW() WHERE id=$1`, id, active)
	if err != nil {
		return Department{}, fmt.Errorf("set department status: %w", err)
	}
	if result.RowsAffected() == 0 {
		return Department{}, ErrNotFound
	}
	return s.Get(ctx, id)
}

func (s *PostgresStore) Delete(ctx context.Context, id string) error {
	var volunteerCount int
	if err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM volunteers WHERE department_id=$1`, id).Scan(&volunteerCount); err != nil {
		return fmt.Errorf("count department volunteers: %w", err)
	}
	if volunteerCount > 0 {
		return ErrInUse
	}
	result, err := s.pool.Exec(ctx, `DELETE FROM departments WHERE id=$1`, id)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23503" {
		return ErrInUse
	}
	if err != nil {
		return fmt.Errorf("delete department: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func mapWriteError(err error) error {
	if err == nil {
		return nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return ErrNameExists
	}
	return fmt.Errorf("write department: %w", err)
}

type scanner interface{ Scan(dest ...any) error }

func scanDepartment(row scanner) (Department, error) {
	var item Department
	if err := row.Scan(&item.ID, &item.Name, &item.Active, &item.VolunteerCount, &item.ActiveVolunteerCount, &item.CreatedAt, &item.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Department{}, ErrNotFound
		}
		return Department{}, fmt.Errorf("scan department: %w", err)
	}
	return item, nil
}
