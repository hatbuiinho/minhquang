package volunteer

import (
	"context"
	"sort"
	"strings"
	"sync"
)

type MemoryStore struct {
	mu    sync.RWMutex
	items map[string]Volunteer
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{items: make(map[string]Volunteer)}
}

func (s *MemoryStore) Create(_ context.Context, item Volunteer) (Volunteer, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[item.ID] = item
	return item, nil
}

func (s *MemoryStore) List(_ context.Context, options ListOptions) ([]Volunteer, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	query := normalizeSearchText(options.Query)
	items := make([]Volunteer, 0, len(s.items))
	for _, item := range s.items {
		departed := item.DepartureDate != nil && item.DepartureDate.Before(options.Today)
		if options.Status == "active" && departed {
			continue
		}
		if options.Status == "departed" && !departed {
			continue
		}
		haystack := normalizeSearchText(item.FullName + " " + item.DharmaName + " " + item.BirthDate + " " + item.Phone + " " + item.CultivationPlace + " " + item.Department + " " + item.Notes)
		if query != "" && !strings.Contains(haystack, query) {
			continue
		}
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].ArrivalDate.Equal(items[j].ArrivalDate) {
			return items[i].FullName < items[j].FullName
		}
		return items[i].ArrivalDate.After(items[j].ArrivalDate)
	})
	return items, nil
}

func (s *MemoryStore) Get(_ context.Context, id string) (Volunteer, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.items[id]
	if !ok {
		return Volunteer{}, ErrNotFound
	}
	return item, nil
}

func (s *MemoryStore) Update(_ context.Context, item Volunteer) (Volunteer, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.items[item.ID]; !ok {
		return Volunteer{}, ErrNotFound
	}
	s.items[item.ID] = item
	return item, nil
}

func (s *MemoryStore) Delete(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.items[id]; !ok {
		return ErrNotFound
	}
	delete(s.items, id)
	return nil
}
