package user

import (
	"context"
	"sort"
	"sync"
	"time"
)

type MemoryStore struct {
	mu       sync.RWMutex
	users    map[string]User
	sessions map[string]Session
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{users: make(map[string]User), sessions: make(map[string]Session)}
}

func (s *MemoryStore) Create(_ context.Context, item User) (User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, existing := range s.users {
		if existing.Username == item.Username {
			return User{}, ErrUsernameExists
		}
	}
	s.users[item.ID] = item
	return item, nil
}

func (s *MemoryStore) List(_ context.Context) ([]User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]User, 0, len(s.users))
	for _, item := range s.users {
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].DisplayName < items[j].DisplayName })
	return items, nil
}

func (s *MemoryStore) FindByUsername(_ context.Context, username string) (User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, item := range s.users {
		if item.Username == username {
			return item, nil
		}
	}
	return User{}, ErrNotFound
}

func (s *MemoryStore) CreateSession(_ context.Context, session Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[session.TokenHash] = session
	return nil
}

func (s *MemoryStore) UserBySession(_ context.Context, tokenHash string) (User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	session, ok := s.sessions[tokenHash]
	if !ok || time.Now().UTC().After(session.ExpiresAt) {
		return User{}, ErrInvalidCredentials
	}
	item, ok := s.users[session.UserID]
	if !ok || !item.Active {
		return User{}, ErrInvalidCredentials
	}
	return item, nil
}

func (s *MemoryStore) DeleteSession(_ context.Context, tokenHash string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, tokenHash)
	return nil
}
