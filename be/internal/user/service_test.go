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
	if created.Role != RoleViewer || !HasPermission(created.Role, PermissionVolunteerRead) || !HasPermission(created.Role, PermissionUserRead) || HasPermission(created.Role, PermissionUserManage) {
		t.Fatalf("unexpected default viewer permissions: %#v", created)
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

func TestCreateAdminReturnsAdminPermissions(t *testing.T) {
	service := NewService(NewMemoryStore(), time.Now)
	created, err := service.Create(context.Background(), CreateInput{
		Username: "admin", DisplayName: "Ban quan tri", Password: "password123", Role: RoleAdmin,
	})
	if err != nil {
		t.Fatalf("create admin: %v", err)
	}
	if created.Role != RoleAdmin || !HasPermission(created.Role, PermissionUserManage) || len(created.Permissions) != 8 {
		t.Fatalf("unexpected admin permissions: %#v", created)
	}
	if _, err := service.Create(context.Background(), CreateInput{
		Username: "invalid", DisplayName: "Invalid", Password: "password123", Role: "owner",
	}); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid role error, got %v", err)
	}
}

func TestUpdateAccountNormalizesFieldsAndRole(t *testing.T) {
	now := time.Date(2026, 8, 11, 10, 0, 0, 0, time.UTC)
	service := NewService(NewMemoryStore(), func() time.Time { return now })
	created, err := service.Create(context.Background(), CreateInput{
		Username: "viewer", DisplayName: "Tai khoan xem", Password: "password123", Role: RoleViewer,
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	actor := User{ID: "admin-id", Role: RoleAdmin}
	updated, err := service.Update(context.Background(), actor, created.ID, UpdateInput{
		Username: "  EDITOR  ", DisplayName: "  Ban Quan Tri  ", Role: RoleAdmin,
	})
	if err != nil {
		t.Fatalf("update user: %v", err)
	}
	if updated.Username != "editor" || updated.DisplayName != "Ban Quan Tri" || updated.Role != RoleAdmin {
		t.Fatalf("unexpected updated user: %#v", updated)
	}
	if !HasPermission(updated.Role, PermissionUserManage) || updated.UpdatedAt != now {
		t.Fatalf("unexpected updated permissions or timestamp: %#v", updated)
	}
	if _, err := service.Update(context.Background(), actor, created.ID, UpdateInput{
		Username: "editor", DisplayName: "Ban Quan Tri", Role: "owner",
	}); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid role error, got %v", err)
	}
	other, err := service.Create(context.Background(), CreateInput{
		Username: "other", DisplayName: "Tai khoan khac", Password: "password123",
	})
	if err != nil {
		t.Fatalf("create other user: %v", err)
	}
	if _, err := service.Update(context.Background(), actor, other.ID, UpdateInput{
		Username: updated.Username, DisplayName: other.DisplayName, Role: other.Role,
	}); !errors.Is(err, ErrUsernameExists) {
		t.Fatalf("expected duplicate username error, got %v", err)
	}
	if _, err := service.Update(context.Background(), actor, "missing", UpdateInput{
		Username: "missing", DisplayName: "Missing", Role: RoleViewer,
	}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected not found error, got %v", err)
	}
	if _, err := service.Update(context.Background(), updated, updated.ID, UpdateInput{
		Username: updated.Username, DisplayName: updated.DisplayName, Role: RoleViewer,
	}); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected self role change error, got %v", err)
	}
}

func TestAdminResetPasswordRevokesTargetSessions(t *testing.T) {
	now := time.Date(2026, 8, 11, 11, 0, 0, 0, time.UTC)
	service := NewService(NewMemoryStore(), func() time.Time { return now })
	target, err := service.Create(context.Background(), CreateInput{
		Username: "viewer", DisplayName: "Giam sat vien", Password: "old-password", Role: RoleViewer,
	})
	if err != nil {
		t.Fatalf("create target: %v", err)
	}
	session, err := service.Login(context.Background(), target.Username, "old-password")
	if err != nil {
		t.Fatalf("login target: %v", err)
	}
	actor := User{ID: "admin-id", Role: RoleAdmin}
	if _, err := service.Update(context.Background(), actor, target.ID, UpdateInput{
		Username: target.Username, DisplayName: target.DisplayName, Role: target.Role, Password: "new-password",
	}); err != nil {
		t.Fatalf("reset target password: %v", err)
	}
	if _, err := service.Authenticate(context.Background(), session.Token); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("target session should be revoked, got %v", err)
	}
	if _, err := service.Login(context.Background(), target.Username, "old-password"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("old password should be rejected, got %v", err)
	}
	if _, err := service.Login(context.Background(), target.Username, "new-password"); err != nil {
		t.Fatalf("new password login: %v", err)
	}
	if _, err := service.Update(context.Background(), target, target.ID, UpdateInput{
		Username: target.Username, DisplayName: target.DisplayName, Role: target.Role, Password: "another-password",
	}); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected self password reset error, got %v", err)
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
