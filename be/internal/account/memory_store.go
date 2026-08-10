package account

import (
	"context"
	"sync"
)

type MemoryStore struct {
	mu           sync.RWMutex
	users        map[string]User
	groups       map[string]Group
	groupMembers map[string]GroupMember
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		users:        make(map[string]User),
		groups:       make(map[string]Group),
		groupMembers: make(map[string]GroupMember),
	}
}

func (s *MemoryStore) CreateUser(_ context.Context, item User) (User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.users[item.ID] = item
	return item, nil
}

func (s *MemoryStore) ListUsers(_ context.Context, activeOnly bool) ([]User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]User, 0, len(s.users))
	for _, item := range s.users {
		if activeOnly && !item.Active {
			continue
		}
		items = append(items, item)
	}
	return items, nil
}

func (s *MemoryStore) CreateGroup(_ context.Context, item Group) (Group, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.groups[item.ID] = item
	return item, nil
}

func (s *MemoryStore) ListGroups(_ context.Context, activeOnly bool) ([]Group, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]Group, 0, len(s.groups))
	for _, item := range s.groups {
		if activeOnly && !item.Active {
			continue
		}
		items = append(items, item)
	}
	return items, nil
}

func (s *MemoryStore) AddGroupMember(_ context.Context, item GroupMember) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.groupMembers[item.GroupID+":"+item.UserID] = item
	return nil
}

func (s *MemoryStore) ListGroupMembers(_ context.Context, groupIDs []string) ([]GroupMember, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	groupSet := make(map[string]struct{}, len(groupIDs))
	for _, groupID := range groupIDs {
		groupSet[groupID] = struct{}{}
	}

	items := make([]GroupMember, 0)
	for _, item := range s.groupMembers {
		if _, ok := groupSet[item.GroupID]; ok {
			items = append(items, item)
		}
	}
	return items, nil
}
