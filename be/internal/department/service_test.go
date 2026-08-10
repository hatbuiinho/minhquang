package department

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestResolveOrCreateReusesUnaccentedName(t *testing.T) {
	service := NewService(NewMemoryStore(), func() time.Time { return time.Date(2026, 8, 10, 0, 0, 0, 0, time.UTC) })
	firstID, firstName, err := service.ResolveOrCreate(context.Background(), " Ban  Bếp ")
	if err != nil {
		t.Fatalf("resolve first department: %v", err)
	}
	secondID, secondName, err := service.ResolveOrCreate(context.Background(), "ban bep")
	if err != nil {
		t.Fatalf("resolve second department: %v", err)
	}
	if firstID != secondID || firstName != "Ban Bếp" || secondName != firstName {
		t.Fatalf("departments were not canonicalized: %q %q / %q %q", firstID, firstName, secondID, secondName)
	}
}

func TestInactiveDepartmentCannotBeResolved(t *testing.T) {
	service := NewService(NewMemoryStore(), time.Now)
	item, err := service.Create(context.Background(), "Ban Bếp")
	if err != nil {
		t.Fatalf("create department: %v", err)
	}
	if _, err := service.SetActive(context.Background(), item.ID, false); err != nil {
		t.Fatalf("deactivate department: %v", err)
	}
	_, _, err = service.ResolveOrCreate(context.Background(), "ban bep")
	if !errors.Is(err, ErrInactive) {
		t.Fatalf("expected inactive error, got %v", err)
	}
}
