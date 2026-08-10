package user

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const sessionDuration = 30 * 24 * time.Hour

type Service struct {
	store Store
	now   func() time.Time
}

func NewService(store Store, now func() time.Time) *Service {
	return &Service{store: store, now: now}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (User, error) {
	username := strings.ToLower(strings.TrimSpace(input.Username))
	displayName := strings.TrimSpace(input.DisplayName)
	if username == "" || displayName == "" {
		return User{}, fmt.Errorf("%w: username and display_name are required", ErrInvalidInput)
	}
	if len(input.Password) < 8 {
		return User{}, fmt.Errorf("%w: password must contain at least 8 characters", ErrInvalidInput)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, fmt.Errorf("hash password: %w", err)
	}
	now := s.now().UTC()
	return s.store.Create(ctx, User{
		ID:           newID("usr"),
		Username:     username,
		DisplayName:  displayName,
		PasswordHash: string(hash),
		Role:         "admin",
		Active:       true,
		CreatedAt:    now,
		UpdatedAt:    now,
	})
}

func (s *Service) EnsureInitialAdmin(ctx context.Context, input CreateInput) error {
	if strings.TrimSpace(input.Username) == "" || input.Password == "" {
		return nil
	}
	if _, err := s.store.FindByUsername(ctx, strings.ToLower(strings.TrimSpace(input.Username))); err == nil {
		return nil
	} else if err != ErrNotFound {
		return err
	}
	_, err := s.Create(ctx, input)
	return err
}

func (s *Service) List(ctx context.Context) ([]User, error) { return s.store.List(ctx) }

func (s *Service) Login(ctx context.Context, username, password string) (LoginResult, error) {
	item, err := s.store.FindByUsername(ctx, strings.ToLower(strings.TrimSpace(username)))
	if err != nil || !item.Active || bcrypt.CompareHashAndPassword([]byte(item.PasswordHash), []byte(password)) != nil {
		return LoginResult{}, ErrInvalidCredentials
	}
	token, err := randomToken()
	if err != nil {
		return LoginResult{}, err
	}
	now := s.now().UTC()
	expiresAt := now.Add(sessionDuration)
	if err := s.store.CreateSession(ctx, Session{TokenHash: hashToken(token), UserID: item.ID, ExpiresAt: expiresAt, CreatedAt: now}); err != nil {
		return LoginResult{}, err
	}
	return LoginResult{Token: token, ExpiresAt: expiresAt, User: item}, nil
}

func (s *Service) Authenticate(ctx context.Context, token string) (User, error) {
	if strings.TrimSpace(token) == "" {
		return User{}, ErrInvalidCredentials
	}
	return s.store.UserBySession(ctx, hashToken(token))
}

func (s *Service) Logout(ctx context.Context, token string) error {
	if strings.TrimSpace(token) == "" {
		return nil
	}
	return s.store.DeleteSession(ctx, hashToken(token))
}

func randomToken() (string, error) {
	var value [32]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", fmt.Errorf("generate session token: %w", err)
	}
	return hex.EncodeToString(value[:]), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func newID(prefix string) string {
	var value [12]byte
	if _, err := rand.Read(value[:]); err != nil {
		return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
	}
	return prefix + "_" + hex.EncodeToString(value[:])
}
