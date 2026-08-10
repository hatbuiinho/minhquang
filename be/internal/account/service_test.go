package account

import (
	"context"
	"testing"
	"time"

	"reminder/be/internal/event"
)

func TestServiceResolveSelectedGroupsDedupesUsers(t *testing.T) {
	now := time.Date(2026, 8, 2, 10, 0, 0, 0, time.UTC)
	service := NewService(NewMemoryStore(), func() time.Time { return now })
	ctx := context.Background()

	user, err := service.CreateUser(ctx, CreateUserInput{Name: "An"})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	firstGroup, err := service.CreateGroup(ctx, CreateGroupInput{Name: "Gia đình"})
	if err != nil {
		t.Fatalf("create first group: %v", err)
	}
	secondGroup, err := service.CreateGroup(ctx, CreateGroupInput{Name: "Công việc"})
	if err != nil {
		t.Fatalf("create second group: %v", err)
	}
	if err := service.AddGroupMember(ctx, firstGroup.ID, user.ID); err != nil {
		t.Fatalf("add first group member: %v", err)
	}
	if err := service.AddGroupMember(ctx, secondGroup.ID, user.ID); err != nil {
		t.Fatalf("add second group member: %v", err)
	}

	recipients, err := service.ResolveEventRecipients(ctx, "owner", event.RecipientInput{
		AudienceType: event.AudienceSelectedGroups,
		GroupIDs:     []string{firstGroup.ID, secondGroup.ID},
	})
	if err != nil {
		t.Fatalf("resolve event recipients: %v", err)
	}
	if len(recipients) != 2 {
		t.Fatalf("resolver preserves source rows for event service dedupe, got %d", len(recipients))
	}
}
