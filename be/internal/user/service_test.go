package user

import (
	"context"
	"errors"
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

func TestChangePasswordKeepsCurrentSessionAndRevokesOthers(t *testing.T) {
	now := time.Date(2026, 8, 10, 9, 0, 0, 0, time.UTC)
	service := NewService(NewMemoryStore(), func() time.Time { return now })
	if _, err := service.Create(context.Background(), CreateInput{
		Username: "admin", DisplayName: "Ban quan tri", Password: "password123",
	}); err != nil {
		t.Fatalf("create user: %v", err)
	}
	currentSession, err := service.Login(context.Background(), "admin", "password123")
	if err != nil {
		t.Fatalf("login current session: %v", err)
	}
	otherSession, err := service.Login(context.Background(), "admin", "password123")
	if err != nil {
		t.Fatalf("login other session: %v", err)
	}
	item, err := service.Authenticate(context.Background(), currentSession.Token)
	if err != nil {
		t.Fatalf("authenticate current session: %v", err)
	}

	err = service.ChangePassword(context.Background(), item, "wrong-password", "new-password-123", currentSession.Token)
	if !errors.Is(err, ErrCurrentPassword) {
		t.Fatalf("expected current password error, got %v", err)
	}
	err = service.ChangePassword(context.Background(), item, "password123", "new-password-123", currentSession.Token)
	if err != nil {
		t.Fatalf("change password: %v", err)
	}
	if _, err := service.Authenticate(context.Background(), currentSession.Token); err != nil {
		t.Fatalf("current session was revoked: %v", err)
	}
	if _, err := service.Authenticate(context.Background(), otherSession.Token); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("other session should be revoked, got %v", err)
	}
	if _, err := service.Login(context.Background(), "admin", "password123"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("old password should be rejected, got %v", err)
	}
	if _, err := service.Login(context.Background(), "admin", "new-password-123"); err != nil {
		t.Fatalf("new password login: %v", err)
	}
}

func TestUpdateProfileNormalizesUsernameAndRejectsDuplicate(t *testing.T) {
	now := time.Date(2026, 8, 10, 9, 0, 0, 0, time.UTC)
	service := NewService(NewMemoryStore(), func() time.Time { return now })
	current, err := service.Create(context.Background(), CreateInput{
		Username: "admin", DisplayName: "Ban quan tri", Password: "password123",
	})
	if err != nil {
		t.Fatalf("create current user: %v", err)
	}
	if _, err := service.Create(context.Background(), CreateInput{
		Username: "other", DisplayName: "Tai khoan khac", Password: "password123",
	}); err != nil {
		t.Fatalf("create other user: %v", err)
	}

	updated, err := service.UpdateProfile(context.Background(), current, "  NEW.ADMIN  ", "  Ban Quan Tri Moi  ")
	if err != nil {
		t.Fatalf("update profile: %v", err)
	}
	if updated.Username != "new.admin" || updated.DisplayName != "Ban Quan Tri Moi" {
		t.Fatalf("unexpected updated profile: %#v", updated)
	}
	if _, err := service.UpdateProfile(context.Background(), updated, "other", updated.DisplayName); !errors.Is(err, ErrUsernameExists) {
		t.Fatalf("expected duplicate username error, got %v", err)
	}
}

func TestUpdateAvatar(t *testing.T) {
	now := time.Date(2026, 8, 11, 9, 0, 0, 0, time.UTC)
	service := NewService(NewMemoryStore(), func() time.Time { return now })
	item, err := service.Create(context.Background(), CreateInput{
		Username: "admin", DisplayName: "Ban quan tri", Password: "password123",
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	updated, err := service.UpdateAvatar(context.Background(), item, " https://media.example.com/avatars/admin.jpg ")
	if err != nil {
		t.Fatalf("update avatar: %v", err)
	}
	if updated.AvatarURL != "https://media.example.com/avatars/admin.jpg" {
		t.Fatalf("unexpected avatar URL %q", updated.AvatarURL)
	}
	if _, err := service.UpdateAvatar(context.Background(), updated, "javascript:alert(1)"); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid avatar URL, got %v", err)
	}
}
