package department

import "context"

type Store interface {
	Create(ctx context.Context, item Department, searchKey string) (Department, error)
	List(ctx context.Context, options ListOptions) ([]Department, error)
	Get(ctx context.Context, id string) (Department, error)
	FindBySearchKey(ctx context.Context, searchKey string) (Department, error)
	Update(ctx context.Context, item Department, searchKey string) (Department, error)
	SetActive(ctx context.Context, id string, active bool) (Department, error)
	Delete(ctx context.Context, id string) error
}
