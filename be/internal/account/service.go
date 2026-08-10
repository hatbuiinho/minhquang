package account

import (
	"context"
	"fmt"
	"strings"
	"time"

	"reminder/be/internal/event"
)

type Service struct {
	store Store
	now   func() time.Time
}

func NewService(store Store, now func() time.Time) *Service {
	return &Service{store: store, now: now}
}

func (s *Service) CreateUser(ctx context.Context, input CreateUserInput) (User, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return User{}, fmt.Errorf("%w: name is required", ErrInvalidInput)
	}

	id := strings.TrimSpace(input.ID)
	if id == "" {
		id = newPrefixedID("usr")
	}
	now := s.now().UTC()
	return s.store.CreateUser(ctx, User{
		ID:        id,
		Name:      name,
		Email:     strings.TrimSpace(input.Email),
		Active:    true,
		CreatedAt: now,
		UpdatedAt: now,
	})
}

func (s *Service) ListUsers(ctx context.Context) ([]User, error) {
	return s.store.ListUsers(ctx, true)
}

func (s *Service) CreateGroup(ctx context.Context, input CreateGroupInput) (Group, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return Group{}, fmt.Errorf("%w: name is required", ErrInvalidInput)
	}

	now := s.now().UTC()
	return s.store.CreateGroup(ctx, Group{
		ID:          newPrefixedID("grp"),
		Name:        name,
		Description: strings.TrimSpace(input.Description),
		Active:      true,
		CreatedAt:   now,
		UpdatedAt:   now,
	})
}

func (s *Service) ListGroups(ctx context.Context) ([]Group, error) {
	return s.store.ListGroups(ctx, true)
}

func (s *Service) AddGroupMember(ctx context.Context, groupID string, userID string) error {
	groupID = strings.TrimSpace(groupID)
	userID = strings.TrimSpace(userID)
	if groupID == "" || userID == "" {
		return fmt.Errorf("%w: group_id and user_id are required", ErrInvalidInput)
	}

	return s.store.AddGroupMember(ctx, GroupMember{
		GroupID:   groupID,
		UserID:    userID,
		CreatedAt: s.now().UTC(),
	})
}

func (s *Service) ResolveEventRecipients(ctx context.Context, ownerUserID string, input event.RecipientInput) ([]event.ResolvedRecipient, error) {
	switch input.AudienceType {
	case event.AudienceSelf, "":
		return []event.ResolvedRecipient{{UserID: ownerUserID, SourceType: event.RecipientSourceSelf}}, nil
	case event.AudienceSelectedUsers:
		return selectedUserRecipients(input.UserIDs), nil
	case event.AudienceSelectedGroups:
		members, err := s.store.ListGroupMembers(ctx, input.GroupIDs)
		if err != nil {
			return nil, err
		}
		recipients := make([]event.ResolvedRecipient, 0, len(members))
		for _, member := range members {
			recipients = append(recipients, event.ResolvedRecipient{
				UserID:     member.UserID,
				SourceType: event.RecipientSourceGroup,
				SourceID:   member.GroupID,
			})
		}
		return recipients, nil
	case event.AudienceAllUsers:
		users, err := s.store.ListUsers(ctx, true)
		if err != nil {
			return nil, err
		}
		recipients := make([]event.ResolvedRecipient, 0, len(users))
		for _, user := range users {
			recipients = append(recipients, event.ResolvedRecipient{
				UserID:     user.ID,
				SourceType: event.RecipientSourceAllUsers,
			})
		}
		return recipients, nil
	default:
		return nil, fmt.Errorf("%w: audience_type is invalid", ErrInvalidInput)
	}
}

func selectedUserRecipients(userIDs []string) []event.ResolvedRecipient {
	recipients := make([]event.ResolvedRecipient, 0, len(userIDs))
	for _, userID := range userIDs {
		userID = strings.TrimSpace(userID)
		if userID == "" {
			continue
		}
		recipients = append(recipients, event.ResolvedRecipient{
			UserID:     userID,
			SourceType: event.RecipientSourceUser,
			SourceID:   userID,
		})
	}
	return recipients
}
