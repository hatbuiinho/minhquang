package user

import "context"

type Store interface {
	Create(ctx context.Context, item User) (User, error)
	List(ctx context.Context) ([]User, error)
	FindByUsername(ctx context.Context, username string) (User, error)
	CreateSession(ctx context.Context, session Session) error
	UserBySession(ctx context.Context, tokenHash string) (User, error)
	DeleteSession(ctx context.Context, tokenHash string) error
}
