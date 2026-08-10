package device

import (
	"context"
	"sync"
)

type Store interface {
	Upsert(ctx context.Context, item Device) (Device, error)
	ListEnabledByUser(ctx context.Context, userID string) ([]Device, error)
}

type MemoryStore struct {
	mu      sync.RWMutex
	devices map[string]Device
	userIDs map[string][]string
	userPos map[string]map[string]int
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		devices: make(map[string]Device),
		userIDs: make(map[string][]string),
		userPos: make(map[string]map[string]int),
	}
}

func (s *MemoryStore) Upsert(_ context.Context, item Device) (Device, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := deviceKey(item.UserID, item.PushToken)
	if current, ok := s.devices[key]; ok {
		item.ID = current.ID
		item.CreatedAt = current.CreatedAt
	}

	s.devices[key] = item
	if s.userPos[item.UserID] == nil {
		s.userPos[item.UserID] = make(map[string]int)
	}
	if _, ok := s.userPos[item.UserID][key]; !ok {
		s.userPos[item.UserID][key] = len(s.userIDs[item.UserID])
		s.userIDs[item.UserID] = append(s.userIDs[item.UserID], key)
	}

	return item, nil
}

func (s *MemoryStore) ListEnabledByUser(_ context.Context, userID string) ([]Device, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	keys := s.userIDs[userID]
	items := make([]Device, 0, len(keys))
	for _, key := range keys {
		item, ok := s.devices[key]
		if ok && item.Enabled {
			items = append(items, item)
		}
	}

	return items, nil
}

func deviceKey(userID string, pushToken string) string {
	return userID + ":" + pushToken
}
