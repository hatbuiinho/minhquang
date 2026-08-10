package user

import (
	"context"
	"testing"
	"time"
)

func TestCreateAndLogin(t *testing.T) {
	now := time.Date(2026, 8, 10, 9, 0, 0, 0, time.UTC)
	service := NewService(NewMemoryStore(), func() time.Time { return now })
	created, err := service.Create(context.Background(), CreateInput{
		Username: "admin", DisplayName: "Ban quan tri", Password: "password123",
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	result, err := service.Login(context.Background(), "ADMIN", "password123")
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if result.Token == "" || result.User.ID != created.ID {
		t.Fatalf("unexpected login result: %#v", result)
	}
	if _, err := service.Authenticate(context.Background(), result.Token); err != nil {
		t.Fatalf("authenticate: %v", err)
	}
}
