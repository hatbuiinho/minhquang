package volunteer

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStore struct{ pool *pgxpool.Pool }

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore { return &PostgresStore{pool: pool} }

const volunteerSelectColumns = `v.id, v.full_name, v.dharma_name, v.birth_date, v.cultivation_place, v.phone, COALESCE(v.department_id, ''), COALESCE(d.name, ''), v.notes, v.avatar_url, v.arrival_date, v.departure_date, v.created_at, v.updated_at`

func (s *PostgresStore) Create(ctx context.Context, item Volunteer) (Volunteer, error) {
	const query = `INSERT INTO volunteers (id, full_name, dharma_name, birth_date, cultivation_place, phone, department_id, notes, avatar_url, arrival_date, departure_date, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,NULLIF($7,''),$8,$9,$10,$11,$12,$13)`
	_, err := s.pool.Exec(ctx, query, item.ID, item.FullName, item.DharmaName, item.BirthDate, item.CultivationPlace, item.Phone, item.DepartmentID, item.Notes, item.AvatarURL, item.ArrivalDate, item.DepartureDate, item.CreatedAt, item.UpdatedAt)
	if err != nil {
		return Volunteer{}, fmt.Errorf("create volunteer: %w", err)
	}
	return item, nil
}

func (s *PostgresStore) List(ctx context.Context, options ListOptions) ([]Volunteer, error) {
	query := `SELECT ` + volunteerSelectColumns + ` FROM volunteers v LEFT JOIN departments d ON d.id=v.department_id WHERE ($1 = '' OR unaccent(lower(concat_ws(' ', v.full_name, v.dharma_name, v.birth_date, v.cultivation_place, v.phone, d.name, v.notes))) LIKE '%' || unaccent(lower($1)) || '%') AND ($2 = '' OR ($2 = 'active' AND (v.departure_date IS NULL OR v.departure_date >= $3)) OR ($2 = 'departed' AND v.departure_date < $3)) ORDER BY v.arrival_date DESC, v.full_name ASC`
	rows, err := s.pool.Query(ctx, query, options.Query, options.Status, options.Today)
	if err != nil {
		return nil, fmt.Errorf("list volunteers: %w", err)
	}
	defer rows.Close()
	items := make([]Volunteer, 0)
	for rows.Next() {
		item, err := scanVolunteer(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *PostgresStore) Get(ctx context.Context, id string) (Volunteer, error) {
	return scanVolunteer(s.pool.QueryRow(ctx, `SELECT `+volunteerSelectColumns+` FROM volunteers v LEFT JOIN departments d ON d.id=v.department_id WHERE v.id=$1`, id))
}

func (s *PostgresStore) Update(ctx context.Context, item Volunteer) (Volunteer, error) {
	const query = `UPDATE volunteers SET full_name=$2, dharma_name=$3, birth_date=$4, cultivation_place=$5, phone=$6, department_id=NULLIF($7,''), notes=$8, avatar_url=$9, arrival_date=$10, departure_date=$11, updated_at=$12 WHERE id=$1`
	result, err := s.pool.Exec(ctx, query, item.ID, item.FullName, item.DharmaName, item.BirthDate, item.CultivationPlace, item.Phone, item.DepartmentID, item.Notes, item.AvatarURL, item.ArrivalDate, item.DepartureDate, item.UpdatedAt)
	if err != nil {
		return Volunteer{}, fmt.Errorf("update volunteer: %w", err)
	}
	if result.RowsAffected() == 0 {
		return Volunteer{}, ErrNotFound
	}
	return item, nil
}

func (s *PostgresStore) Delete(ctx context.Context, id string) error {
	result, err := s.pool.Exec(ctx, `DELETE FROM volunteers WHERE id=$1`, id)
	if err != nil {
		return fmt.Errorf("delete volunteer: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

type scanner interface{ Scan(dest ...any) error }

func scanVolunteer(row scanner) (Volunteer, error) {
	var item Volunteer
	if err := row.Scan(&item.ID, &item.FullName, &item.DharmaName, &item.BirthDate, &item.CultivationPlace, &item.Phone, &item.DepartmentID, &item.Department, &item.Notes, &item.AvatarURL, &item.ArrivalDate, &item.DepartureDate, &item.CreatedAt, &item.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Volunteer{}, ErrNotFound
		}
		return Volunteer{}, fmt.Errorf("scan volunteer: %w", err)
	}
	return item, nil
}
