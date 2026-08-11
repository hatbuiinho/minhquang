package volunteer

import "context"

type Store interface {
	Create(ctx context.Context, item Volunteer) (Volunteer, error)
	List(ctx context.Context, options ListOptions) ([]Volunteer, error)
	Count(ctx context.Context, options ListOptions) (int, error)
	Get(ctx context.Context, id string) (Volunteer, error)
	Update(ctx context.Context, item Volunteer) (Volunteer, error)
	Delete(ctx context.Context, id string) error
	BulkUpdate(ctx context.Context, ids []string, patch BulkPatch) (int, error)
	BulkDelete(ctx context.Context, ids []string) (int, error)
}

type DepartmentResolver interface {
	ResolveOrCreate(ctx context.Context, name string) (id string, canonicalName string, err error)
}
