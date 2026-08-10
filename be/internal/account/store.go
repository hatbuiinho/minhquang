package account

import "context"

type Store interface {
	CreateUser(ctx context.Context, item User) (User, error)
	ListUsers(ctx context.Context, activeOnly bool) ([]User, error)
	CreateGroup(ctx context.Context, item Group) (Group, error)
	ListGroups(ctx context.Context, activeOnly bool) ([]Group, error)
	AddGroupMember(ctx context.Context, item GroupMember) error
	ListGroupMembers(ctx context.Context, groupIDs []string) ([]GroupMember, error)
}
