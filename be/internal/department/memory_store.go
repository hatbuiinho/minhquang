package department

import (
	"context"
	"sort"
	"strings"
	"sync"
)

type MemoryStore struct {
	mu         sync.RWMutex
	items      map[string]Department
	searchKeys map[string]string
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{items: make(map[string]Department), searchKeys: make(map[string]string)}
}

func (s *MemoryStore) Create(_ context.Context, item Department, key string) (Department, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.searchKeys[key]; exists {
		return Department{}, ErrNameExists
	}
	s.items[item.ID] = item
	s.searchKeys[key] = item.ID
	return item, nil
}

func (s *MemoryStore) List(_ context.Context, options ListOptions) ([]Department, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	query := normalizeSearch(options.Query)
	items := make([]Department, 0, len(s.items))
	for _, item := range s.items {
		if options.Active != nil && item.Active != *options.Active {
			continue
		}
		if query != "" && !strings.Contains(normalizeSearch(item.Name), query) {
			continue
		}
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].Active != items[j].Active {
			return items[i].Active
		}
		return items[i].Name < items[j].Name
	})
	if options.Limit > 0 && len(items) > options.Limit {
		items = items[:options.Limit]
	}
	return items, nil
}

func (s *MemoryStore) Get(_ context.Context, id string) (Department, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.items[id]
	if !ok {
		return Department{}, ErrNotFound
	}
	return item, nil
}

func (s *MemoryStore) FindBySearchKey(_ context.Context, key string) (Department, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	id, ok := s.searchKeys[key]
	if !ok {
		return Department{}, ErrNotFound
	}
	return s.items[id], nil
}

func (s *MemoryStore) Update(_ context.Context, item Department, key string) (Department, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	current, ok := s.items[item.ID]
	if !ok {
		return Department{}, ErrNotFound
	}
	if existingID, exists := s.searchKeys[key]; exists && existingID != item.ID {
		return Department{}, ErrNameExists
	}
	delete(s.searchKeys, normalizeSearch(current.Name))
	s.searchKeys[key] = item.ID
	s.items[item.ID] = item
	return item, nil
}

func (s *MemoryStore) SetActive(_ context.Context, id string, active bool) (Department, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.items[id]
	if !ok {
		return Department{}, ErrNotFound
	}
	item.Active = active
	s.items[id] = item
	return item, nil
}

func (s *MemoryStore) Delete(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.items[id]
	if !ok {
		return ErrNotFound
	}
	delete(s.searchKeys, normalizeSearch(item.Name))
	delete(s.items, id)
	return nil
}
