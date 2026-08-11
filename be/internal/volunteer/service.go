package volunteer

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

type Service struct {
	store              Store
	departmentResolver DepartmentResolver
	now                func() time.Time
}

func (s *Service) SetDepartmentResolver(resolver DepartmentResolver) {
	s.departmentResolver = resolver
}

func NewService(store Store, now func() time.Time) *Service {
	return &Service{store: store, now: now}
}

func (s *Service) Create(ctx context.Context, input Input) (Volunteer, error) {
	item, err := s.normalize(input)
	if err != nil {
		return Volunteer{}, err
	}
	if err := s.resolveDepartment(ctx, &item); err != nil {
		return Volunteer{}, err
	}
	now := s.now().UTC()
	item.ID = newID()
	item.CreatedAt = now
	item.UpdatedAt = now
	created, err := s.store.Create(ctx, item)
	if err != nil {
		return Volunteer{}, err
	}
	return s.withStatus(created), nil
}

func (s *Service) List(ctx context.Context, options ListOptions) ([]Volunteer, error) {
	options.Query = strings.TrimSpace(options.Query)
	options.Status = strings.TrimSpace(options.Status)
	options.DepartmentID = strings.TrimSpace(options.DepartmentID)
	options.SortBy = strings.TrimSpace(options.SortBy)
	options.SortDirection = strings.ToLower(strings.TrimSpace(options.SortDirection))
	if options.Status != "" && options.Status != "active" && options.Status != "departed" {
		return nil, fmt.Errorf("%w: status must be active or departed", ErrInvalidInput)
	}
	if options.Limit < 0 || options.Offset < 0 {
		return nil, fmt.Errorf("%w: limit and offset must not be negative", ErrInvalidInput)
	}
	if options.SortBy == "" {
		options.SortBy = "arrival_date"
	}
	if !validSortColumn(options.SortBy) {
		return nil, fmt.Errorf("%w: unsupported sort column", ErrInvalidInput)
	}
	if options.SortDirection == "" {
		options.SortDirection = "desc"
	}
	if options.SortDirection != "asc" && options.SortDirection != "desc" {
		return nil, fmt.Errorf("%w: sort direction must be asc or desc", ErrInvalidInput)
	}
	options.Today = s.today()
	items, err := s.store.List(ctx, options)
	if err != nil {
		return nil, err
	}
	for index := range items {
		items[index] = s.withStatus(items[index])
	}
	return items, nil
}

func validSortColumn(value string) bool {
	switch value {
	case "full_name", "dharma_name", "birth_date", "cultivation_place", "department", "phone", "arrival_date", "departure_date", "status":
		return true
	default:
		return false
	}
}

func (s *Service) Count(ctx context.Context, options ListOptions) (int, error) {
	options.Query = strings.TrimSpace(options.Query)
	options.Status = strings.TrimSpace(options.Status)
	options.DepartmentID = strings.TrimSpace(options.DepartmentID)
	if options.Status != "" && options.Status != "active" && options.Status != "departed" {
		return 0, fmt.Errorf("%w: status must be active or departed", ErrInvalidInput)
	}
	options.Today = s.today()
	return s.store.Count(ctx, options)
}

func (s *Service) Get(ctx context.Context, id string) (Volunteer, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return Volunteer{}, fmt.Errorf("%w: id is required", ErrInvalidInput)
	}
	item, err := s.store.Get(ctx, id)
	if err != nil {
		return Volunteer{}, err
	}
	return s.withStatus(item), nil
}

func (s *Service) Update(ctx context.Context, id string, input Input) (Volunteer, error) {
	current, err := s.Get(ctx, id)
	if err != nil {
		return Volunteer{}, err
	}
	item, err := s.normalize(input)
	if err != nil {
		return Volunteer{}, err
	}
	if err := s.resolveDepartment(ctx, &item); err != nil {
		return Volunteer{}, err
	}
	item.ID = current.ID
	item.CreatedAt = current.CreatedAt
	item.UpdatedAt = s.now().UTC()
	updated, err := s.store.Update(ctx, item)
	if err != nil {
		return Volunteer{}, err
	}
	return s.withStatus(updated), nil
}

func (s *Service) resolveDepartment(ctx context.Context, item *Volunteer) error {
	if s.departmentResolver == nil {
		return nil
	}
	id, name, err := s.departmentResolver.ResolveOrCreate(ctx, item.Department)
	if err != nil {
		return fmt.Errorf("%w: department cannot be used: %v", ErrInvalidInput, err)
	}
	item.DepartmentID = id
	item.Department = name
	return nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return fmt.Errorf("%w: id is required", ErrInvalidInput)
	}
	return s.store.Delete(ctx, id)
}

func (s *Service) BulkDelete(ctx context.Context, ids []string) (int, error) {
	ids, err := normalizeBulkIDs(ids)
	if err != nil {
		return 0, err
	}
	return s.store.BulkDelete(ctx, ids)
}

func (s *Service) BulkUpdate(ctx context.Context, ids []string, field, value string) (int, error) {
	ids, err := normalizeBulkIDs(ids)
	if err != nil {
		return 0, err
	}
	patch := BulkPatch{Field: strings.TrimSpace(field), UpdatedAt: s.now().UTC()}
	value = strings.TrimSpace(value)
	switch patch.Field {
	case "full_name":
		if value == "" {
			return 0, fmt.Errorf("%w: full_name is required", ErrInvalidInput)
		}
		patch.TextValue = value
	case "dharma_name", "birth_date", "cultivation_place", "phone", "notes", "avatar_url":
		patch.TextValue = value
	case "department":
		if utf8.RuneCountInString(strings.Join(strings.Fields(value), " ")) > 60 {
			return 0, fmt.Errorf("%w: department must not exceed 60 characters", ErrInvalidInput)
		}
		patch.Department = strings.Join(strings.Fields(value), " ")
		if s.departmentResolver != nil {
			id, name, err := s.departmentResolver.ResolveOrCreate(ctx, patch.Department)
			if err != nil {
				return 0, fmt.Errorf("%w: department cannot be used: %v", ErrInvalidInput, err)
			}
			patch.DepartmentID, patch.Department = id, name
		}
	case "arrival_date":
		date, err := time.Parse("2006-01-02", value)
		if err != nil {
			return 0, fmt.Errorf("%w: arrival_date must be a valid date", ErrInvalidInput)
		}
		date = dateOnly(date)
		patch.DateValue = &date
	case "departure_date":
		if value != "" {
			date, err := time.Parse("2006-01-02", value)
			if err != nil {
				return 0, fmt.Errorf("%w: departure_date must be a valid date", ErrInvalidInput)
			}
			date = dateOnly(date)
			patch.DateValue = &date
		}
	default:
		return 0, fmt.Errorf("%w: unsupported bulk update field", ErrInvalidInput)
	}
	return s.store.BulkUpdate(ctx, ids, patch)
}

func normalizeBulkIDs(ids []string) ([]string, error) {
	if len(ids) == 0 {
		return nil, fmt.Errorf("%w: at least one id is required", ErrInvalidInput)
	}
	if len(ids) > 500 {
		return nil, fmt.Errorf("%w: no more than 500 ids are allowed", ErrInvalidInput)
	}
	unique := make([]string, 0, len(ids))
	seen := make(map[string]struct{}, len(ids))
	for _, rawID := range ids {
		id := strings.TrimSpace(rawID)
		if id == "" {
			return nil, fmt.Errorf("%w: ids must not be empty", ErrInvalidInput)
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		unique = append(unique, id)
	}
	return unique, nil
}

func (s *Service) normalize(input Input) (Volunteer, error) {
	fullName := strings.TrimSpace(input.FullName)
	if fullName == "" {
		return Volunteer{}, fmt.Errorf("%w: full_name is required", ErrInvalidInput)
	}
	if input.ArrivalDate.IsZero() {
		return Volunteer{}, fmt.Errorf("%w: arrival_date is required", ErrInvalidInput)
	}
	arrival := dateOnly(input.ArrivalDate)
	department := strings.Join(strings.Fields(input.Department), " ")
	if utf8.RuneCountInString(department) > 60 {
		return Volunteer{}, fmt.Errorf("%w: department must not exceed 60 characters", ErrInvalidInput)
	}
	var departure *time.Time
	if input.DepartureDate != nil {
		value := dateOnly(*input.DepartureDate)
		if value.Before(arrival) {
			return Volunteer{}, fmt.Errorf("%w: departure_date cannot be before arrival_date", ErrInvalidInput)
		}
		departure = &value
	}
	return Volunteer{
		FullName:         fullName,
		DharmaName:       strings.TrimSpace(input.DharmaName),
		BirthDate:        strings.TrimSpace(input.BirthDate),
		CultivationPlace: strings.TrimSpace(input.CultivationPlace),
		Phone:            strings.TrimSpace(input.Phone),
		Department:       department,
		Notes:            strings.TrimSpace(input.Notes),
		AvatarURL:        strings.TrimSpace(input.AvatarURL),
		ArrivalDate:      arrival,
		DepartureDate:    departure,
	}, nil
}

func dateOnly(value time.Time) time.Time {
	value = value.UTC()
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
}

func (s *Service) today() time.Time {
	location := time.FixedZone("Asia/Ho_Chi_Minh", 7*60*60)
	now := s.now().In(location)
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
}

func (s *Service) withStatus(item Volunteer) Volunteer {
	item.Status = StatusActive
	if item.DepartureDate != nil && item.DepartureDate.Before(s.today()) {
		item.Status = StatusDeparted
	}
	return item
}

func newID() string {
	var value [12]byte
	if _, err := rand.Read(value[:]); err != nil {
		return fmt.Sprintf("vol_%d", time.Now().UnixNano())
	}
	return "vol_" + hex.EncodeToString(value[:])
}
