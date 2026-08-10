package device

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type Service struct {
	store Store
	now   func() time.Time
}

func NewService(store Store, now func() time.Time) *Service {
	return &Service{store: store, now: now}
}

func (s *Service) Register(ctx context.Context, input RegisterInput) (Device, error) {
	userID := strings.TrimSpace(input.UserID)
	pushToken := strings.TrimSpace(input.PushToken)
	platform := input.Platform

	if userID == "" {
		return Device{}, fmt.Errorf("%w: user_id is required", ErrInvalidInput)
	}
	if pushToken == "" {
		return Device{}, fmt.Errorf("%w: push_token is required", ErrInvalidInput)
	}
	if !validPlatform(platform) {
		return Device{}, fmt.Errorf("%w: platform is invalid", ErrInvalidInput)
	}

	now := s.now().UTC()
	return s.store.Upsert(ctx, Device{
		ID:         newID(),
		UserID:     userID,
		Platform:   platform,
		PushToken:  pushToken,
		Enabled:    true,
		LastSeenAt: now,
		CreatedAt:  now,
		UpdatedAt:  now,
	})
}

func (s *Service) ListEnabledByUser(ctx context.Context, userID string) ([]Device, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, fmt.Errorf("%w: user_id is required", ErrInvalidInput)
	}

	return s.store.ListEnabledByUser(ctx, userID)
}

func validPlatform(platform Platform) bool {
	switch platform {
	case PlatformAndroid, PlatformIOS, PlatformWeb:
		return true
	default:
		return false
	}
}
