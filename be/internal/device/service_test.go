package device

import (
	"context"
	"testing"
	"time"
)

func TestServiceRegisterUpsertsDeviceToken(t *testing.T) {
	now := time.Date(2026, 8, 2, 10, 0, 0, 0, time.UTC)
	service := NewService(NewMemoryStore(), func() time.Time { return now })
	ctx := context.Background()

	created, err := service.Register(ctx, RegisterInput{
		UserID:    "user_1",
		Platform:  PlatformAndroid,
		PushToken: "token_1",
	})
	if err != nil {
		t.Fatalf("register device: %v", err)
	}

	now = now.Add(time.Minute)
	updated, err := service.Register(ctx, RegisterInput{
		UserID:    "user_1",
		Platform:  PlatformAndroid,
		PushToken: "token_1",
	})
	if err != nil {
		t.Fatalf("register device again: %v", err)
	}

	if updated.ID != created.ID {
		t.Fatalf("expected upsert to keep device id, got %q and %q", created.ID, updated.ID)
	}
	if !updated.LastSeenAt.After(created.LastSeenAt) {
		t.Fatalf("expected last_seen_at to update, got %v and %v", created.LastSeenAt, updated.LastSeenAt)
	}

	items, err := service.ListEnabledByUser(ctx, "user_1")
	if err != nil {
		t.Fatalf("list enabled devices: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected one device, got %d", len(items))
	}
}

func TestServiceRegisterRejectsInvalidInput(t *testing.T) {
	service := NewService(NewMemoryStore(), time.Now)
	_, err := service.Register(context.Background(), RegisterInput{
		UserID:    "user_1",
		Platform:  Platform("desktop"),
		PushToken: "token_1",
	})
	if err == nil {
		t.Fatal("expected invalid input error")
	}
}
