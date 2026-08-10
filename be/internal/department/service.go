package department

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

type Service struct {
	store Store
	now   func() time.Time
}

func NewService(store Store, now func() time.Time) *Service {
	return &Service{store: store, now: now}
}

func (s *Service) Create(ctx context.Context, name string) (Department, error) {
	name, key, err := normalizeName(name)
	if err != nil {
		return Department{}, err
	}
	now := s.now().UTC()
	return s.store.Create(ctx, Department{ID: newID(), Name: name, Active: true, CreatedAt: now, UpdatedAt: now}, key)
}

func (s *Service) List(ctx context.Context, options ListOptions) ([]Department, error) {
	options.Query = strings.TrimSpace(options.Query)
	if options.Limit < 0 {
		return nil, fmt.Errorf("%w: limit must not be negative", ErrInvalidInput)
	}
	if options.Limit > 100 {
		options.Limit = 100
	}
	return s.store.List(ctx, options)
}

func (s *Service) Update(ctx context.Context, id, name string) (Department, error) {
	item, err := s.store.Get(ctx, strings.TrimSpace(id))
	if err != nil {
		return Department{}, err
	}
	name, key, err := normalizeName(name)
	if err != nil {
		return Department{}, err
	}
	item.Name = name
	item.UpdatedAt = s.now().UTC()
	return s.store.Update(ctx, item, key)
}

func (s *Service) SetActive(ctx context.Context, id string, active bool) (Department, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return Department{}, fmt.Errorf("%w: id is required", ErrInvalidInput)
	}
	return s.store.SetActive(ctx, id, active)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return fmt.Errorf("%w: id is required", ErrInvalidInput)
	}
	return s.store.Delete(ctx, id)
}

// ResolveOrCreate keeps the volunteer form free-text while preserving one canonical department.
func (s *Service) ResolveOrCreate(ctx context.Context, rawName string) (string, string, error) {
	if strings.TrimSpace(rawName) == "" {
		return "", "", nil
	}
	name, key, err := normalizeName(rawName)
	if err != nil {
		return "", "", err
	}
	item, err := s.store.FindBySearchKey(ctx, key)
	if err == nil {
		if !item.Active {
			return "", "", ErrInactive
		}
		return item.ID, item.Name, nil
	}
	if !errors.Is(err, ErrNotFound) {
		return "", "", err
	}
	created, err := s.Create(ctx, name)
	if errors.Is(err, ErrNameExists) {
		created, err = s.store.FindBySearchKey(ctx, key)
	}
	if err != nil {
		return "", "", err
	}
	if !created.Active {
		return "", "", ErrInactive
	}
	return created.ID, created.Name, nil
}

func normalizeName(value string) (string, string, error) {
	name := strings.Join(strings.Fields(value), " ")
	if name == "" {
		return "", "", fmt.Errorf("%w: name is required", ErrInvalidInput)
	}
	if utf8.RuneCountInString(name) > 60 {
		return "", "", fmt.Errorf("%w: name must not exceed 60 characters", ErrInvalidInput)
	}
	return name, normalizeSearch(name), nil
}

func normalizeSearch(value string) string {
	var result strings.Builder
	for _, character := range norm.NFD.String(strings.ToLower(value)) {
		if unicode.Is(unicode.Mn, character) {
			continue
		}
		if character == 'đ' {
			character = 'd'
		}
		result.WriteRune(character)
	}
	return result.String()
}

func newID() string {
	var value [12]byte
	if _, err := rand.Read(value[:]); err != nil {
		return fmt.Sprintf("dep_%d", time.Now().UnixNano())
	}
	return "dep_" + hex.EncodeToString(value[:])
}
