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
		if !matchesVolunteer(item, options, query) {
			continue
		}
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool { return volunteerLess(items[i], items[j], options) })
	start := min(options.Offset, len(items))
	end := len(items)
	if options.Limit > 0 {
		end = min(start+options.Limit, len(items))
	}
	items = items[start:end]
	return items, nil
}

func volunteerLess(left, right Volunteer, options ListOptions) bool {
	leftValue, rightValue := volunteerSortValue(left, options), volunteerSortValue(right, options)
	if leftValue == rightValue {
		return left.ID < right.ID
	}
	if leftValue == "" {
		return false
	}
	if rightValue == "" {
		return true
	}
	if options.SortDirection == "desc" {
		return leftValue > rightValue
	}
	return leftValue < rightValue
}

func volunteerSortValue(item Volunteer, options ListOptions) string {
	switch options.SortBy {
	case "full_name":
		return normalizeSearchText(item.FullName)
	case "dharma_name":
		return normalizeSearchText(item.DharmaName)
	case "birth_date":
		return normalizeSearchText(item.BirthDate)
	case "cultivation_place":
		return normalizeSearchText(item.CultivationPlace)
	case "department":
		return normalizeSearchText(item.Department)
	case "phone":
		return item.Phone
	case "departure_date":
		if item.DepartureDate == nil {
			return ""
		}
		return item.DepartureDate.Format("2006-01-02")
	case "status":
		if item.DepartureDate != nil && item.DepartureDate.Before(options.Today) {
			return string(StatusDeparted)
		}
		return string(StatusActive)
	default:
		return item.ArrivalDate.Format("2006-01-02")
	}
}

func (s *MemoryStore) Count(_ context.Context, options ListOptions) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	query := normalizeSearchText(options.Query)
	total := 0
	for _, item := range s.items {
		if matchesVolunteer(item, options, query) {
			total++
		}
	}
	return total, nil
}

func matchesVolunteer(item Volunteer, options ListOptions, query string) bool {
	if options.DepartmentID != "" && item.DepartmentID != options.DepartmentID {
		return false
	}
	departed := item.DepartureDate != nil && item.DepartureDate.Before(options.Today)
	if options.Status == "active" && departed {
		return false
	}
	if options.Status == "departed" && !departed {
		return false
	}
	haystack := normalizeSearchText(item.FullName + " " + item.DharmaName + " " + item.BirthDate + " " + item.Phone + " " + item.CultivationPlace + " " + item.Department + " " + item.Notes)
	return query == "" || strings.Contains(haystack, query)
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
