package volunteer

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
	orderColumns := map[string]string{
		"full_name": "v.full_name", "dharma_name": "v.dharma_name", "birth_date": "v.birth_date",
		"cultivation_place": "v.cultivation_place", "department": "COALESCE(d.name, '')", "phone": "v.phone",
		"arrival_date": "v.arrival_date", "departure_date": "v.departure_date",
		"status": "CASE WHEN v.departure_date IS NOT NULL AND v.departure_date < $3 THEN 'departed' ELSE 'active' END",
	}
	direction := "ASC"
	if options.SortDirection == "desc" {
		direction = "DESC"
	}
	query := fmt.Sprintf(`SELECT %s FROM volunteers v LEFT JOIN departments d ON d.id=v.department_id WHERE ($1 = '' OR unaccent(lower(concat_ws(' ', v.full_name, v.dharma_name, v.birth_date, v.cultivation_place, v.phone, d.name, v.notes))) LIKE '%%' || unaccent(lower($1)) || '%%') AND ($2 = '' OR ($2 = 'active' AND (v.departure_date IS NULL OR v.departure_date >= $3)) OR ($2 = 'departed' AND v.departure_date < $3)) AND ($4 = '' OR v.department_id=$4) ORDER BY %s %s NULLS LAST, v.id ASC LIMIT NULLIF($5, 0) OFFSET $6`, volunteerSelectColumns, orderColumns[options.SortBy], direction)
	rows, err := s.pool.Query(ctx, query, options.Query, options.Status, options.Today, options.DepartmentID, options.Limit, options.Offset)
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

func (s *PostgresStore) Count(ctx context.Context, options ListOptions) (int, error) {
	const query = `SELECT COUNT(*) FROM volunteers v LEFT JOIN departments d ON d.id=v.department_id WHERE ($1 = '' OR unaccent(lower(concat_ws(' ', v.full_name, v.dharma_name, v.birth_date, v.cultivation_place, v.phone, d.name, v.notes))) LIKE '%' || unaccent(lower($1)) || '%') AND ($2 = '' OR ($2 = 'active' AND (v.departure_date IS NULL OR v.departure_date >= $3)) OR ($2 = 'departed' AND v.departure_date < $3)) AND ($4 = '' OR v.department_id=$4)`
	var total int
	if err := s.pool.QueryRow(ctx, query, options.Query, options.Status, options.Today, options.DepartmentID).Scan(&total); err != nil {
		return 0, fmt.Errorf("count volunteers: %w", err)
	}
	return total, nil
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

func (s *PostgresStore) BulkUpdate(ctx context.Context, ids []string, patch BulkPatch) (int, error) {
	assignments := map[string]string{
		"full_name": "full_name=$2", "dharma_name": "dharma_name=$2", "birth_date": "birth_date=$2",
		"cultivation_place": "cultivation_place=$2", "phone": "phone=$2", "notes": "notes=$2",
		"avatar_url": "avatar_url=$2", "department": "department_id=NULLIF($2,'')",
		"arrival_date": "arrival_date=$2", "departure_date": "departure_date=$2",
	}
	assignment, ok := assignments[patch.Field]
	if !ok {
		return 0, fmt.Errorf("%w: unsupported bulk update field", ErrInvalidInput)
	}
	var value any = patch.TextValue
	switch patch.Field {
	case "department":
		value = patch.DepartmentID
	case "arrival_date", "departure_date":
		if patch.DateValue == nil {
			value = nil
		} else {
			value = *patch.DateValue
		}
	}
	result, err := s.pool.Exec(ctx, `UPDATE volunteers SET `+assignment+`, updated_at=$3 WHERE id=ANY($1)`, ids, value, patch.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23514" {
			return 0, fmt.Errorf("%w: date conflicts with an existing arrival or departure date", ErrInvalidInput)
		}
		return 0, fmt.Errorf("bulk update volunteers: %w", err)
	}
	return int(result.RowsAffected()), nil
}

func (s *PostgresStore) BulkDelete(ctx context.Context, ids []string) (int, error) {
	result, err := s.pool.Exec(ctx, `DELETE FROM volunteers WHERE id=ANY($1)`, ids)
	if err != nil {
		return 0, fmt.Errorf("bulk delete volunteers: %w", err)
	}
	return int(result.RowsAffected()), nil
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
