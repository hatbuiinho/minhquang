package user

import (
	"context"
	"time"
)

type Store interface {
	Create(ctx context.Context, item User) (User, error)
	List(ctx context.Context) ([]User, error)
	UpdateAccount(ctx context.Context, id, username, displayName, role, passwordHash string, updatedAt time.Time) (User, error)
	FindByUsername(ctx context.Context, username string) (User, error)
	CreateSession(ctx context.Context, session Session) error
	UserBySession(ctx context.Context, tokenHash string) (User, error)
	DeleteSession(ctx context.Context, tokenHash string) error
	ChangePassword(ctx context.Context, userID, passwordHash string, updatedAt time.Time, keepTokenHash string) error
	UpdateProfile(ctx context.Context, item User) (User, error)
	UpdateAvatar(ctx context.Context, item User) (User, error)
}
