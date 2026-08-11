package volunteer

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestCreateAndListVolunteer(t *testing.T) {
	now := time.Date(2026, 8, 10, 9, 0, 0, 0, time.UTC)
	service := NewService(NewMemoryStore(), func() time.Time { return now })
	created, err := service.Create(context.Background(), Input{
		FullName:    "Nguyen Van An",
		DharmaName:  "Tam An",
		BirthDate:   "Khoảng năm 1995",
		ArrivalDate: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("create volunteer: %v", err)
	}
	if created.ID == "" || created.FullName != "Nguyen Van An" {
		t.Fatalf("unexpected volunteer: %#v", created)
	}
	if created.BirthDate != "Khoảng năm 1995" {
		t.Fatalf("unexpected birth date: %q", created.BirthDate)
	}

	items, err := service.List(context.Background(), ListOptions{Query: "tam an", Status: "active"})
	if err != nil {
		t.Fatalf("list volunteers: %v", err)
	}
	if len(items) != 1 || items[0].ID != created.ID {
		t.Fatalf("unexpected list: %#v", items)
	}
}

func TestDepartureCannotPrecedeArrival(t *testing.T) {
	now := time.Date(2026, 8, 10, 9, 0, 0, 0, time.UTC)
	service := NewService(NewMemoryStore(), func() time.Time { return now })
	departure := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	_, err := service.Create(context.Background(), Input{
		FullName:      "Nguyen Van An",
		ArrivalDate:   time.Date(2026, 8, 2, 0, 0, 0, 0, time.UTC),
		DepartureDate: &departure,
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid input, got %v", err)
	}
}

func TestStatusFollowsDepartureDateInVietnam(t *testing.T) {
	// 18:00 UTC is already the next calendar day in Vietnam.
	now := time.Date(2026, 8, 10, 18, 0, 0, 0, time.UTC)
	service := NewService(NewMemoryStore(), func() time.Time { return now })

	yesterday := time.Date(2026, 8, 10, 0, 0, 0, 0, time.UTC)
	today := time.Date(2026, 8, 11, 0, 0, 0, 0, time.UTC)
	tomorrow := time.Date(2026, 8, 12, 0, 0, 0, 0, time.UTC)
	cases := []struct {
		name      string
		departure *time.Time
		want      Status
	}{
		{name: "no departure date", want: StatusActive},
		{name: "past departure date", departure: &yesterday, want: StatusDeparted},
		{name: "departure today", departure: &today, want: StatusActive},
		{name: "future departure date", departure: &tomorrow, want: StatusActive},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			item, err := service.Create(context.Background(), Input{
				FullName:      test.name,
				ArrivalDate:   time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
				DepartureDate: test.departure,
			})
			if err != nil {
				t.Fatalf("create volunteer: %v", err)
			}
			if item.Status != test.want {
				t.Fatalf("status = %q, want %q", item.Status, test.want)
			}
		})
	}

	departed, err := service.List(context.Background(), ListOptions{Status: "departed"})
	if err != nil {
		t.Fatalf("list departed: %v", err)
	}
	if len(departed) != 1 || departed[0].Status != StatusDeparted {
		t.Fatalf("unexpected departed list: %#v", departed)
	}

	active, err := service.List(context.Background(), ListOptions{Status: "active"})
	if err != nil {
		t.Fatalf("list active: %v", err)
	}
	if len(active) != 3 {
		t.Fatalf("active count = %d, want 3", len(active))
	}
}

func TestSearchIgnoresVietnameseAccents(t *testing.T) {
	now := time.Date(2026, 8, 10, 9, 0, 0, 0, time.UTC)
	service := NewService(NewMemoryStore(), func() time.Time { return now })
	_, err := service.Create(context.Background(), Input{
		FullName:    "NGUYỄN VĂN ĐẠT",
		DharmaName:  "TRÍ ĐỨC",
		ArrivalDate: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("create volunteer: %v", err)
	}

	for _, query := range []string{"nguyen van dat", "tri duc"} {
		items, err := service.List(context.Background(), ListOptions{Query: query})
		if err != nil {
			t.Fatalf("search %q: %v", query, err)
		}
		if len(items) != 1 {
			t.Fatalf("search %q returned %d items, want 1", query, len(items))
		}
	}
}

func TestDepartmentIsLimitedToSixtyCharacters(t *testing.T) {
	now := time.Date(2026, 8, 10, 9, 0, 0, 0, time.UTC)
	service := NewService(NewMemoryStore(), func() time.Time { return now })
	_, err := service.Create(context.Background(), Input{
		FullName:    "Nguyen Van An",
		Department:  strings.Repeat("a", 61),
		ArrivalDate: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid input, got %v", err)
	}
}

func TestBulkUpdateAndDeleteVolunteers(t *testing.T) {
	now := time.Date(2026, 8, 10, 9, 0, 0, 0, time.UTC)
	store := NewMemoryStore()
	service := NewService(store, func() time.Time { return now })
	ids := make([]string, 0, 2)
	for _, name := range []string{"Nguyen Van An", "Tran Van Binh"} {
		item, err := service.Create(context.Background(), Input{
			FullName:    name,
			ArrivalDate: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
		})
		if err != nil {
			t.Fatalf("create volunteer: %v", err)
		}
		ids = append(ids, item.ID)
	}

	updated, err := service.BulkUpdate(context.Background(), ids, "departure_date", "2026-08-09")
	if err != nil || updated != 2 {
		t.Fatalf("bulk update = %d, %v; want 2, nil", updated, err)
	}
	departed, err := service.List(context.Background(), ListOptions{Status: "departed"})
	if err != nil || len(departed) != 2 {
		t.Fatalf("departed volunteers = %d, %v; want 2, nil", len(departed), err)
	}

	_, err = service.BulkUpdate(context.Background(), ids, "arrival_date", "2026-08-10")
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected atomic date validation error, got %v", err)
	}
	for _, id := range ids {
		item, getErr := service.Get(context.Background(), id)
		if getErr != nil || item.ArrivalDate.Format("2006-01-02") != "2026-08-01" {
			t.Fatalf("arrival date changed after failed bulk update: %#v, %v", item, getErr)
		}
	}

	deleted, err := service.BulkDelete(context.Background(), append(ids, ids[0]))
	if err != nil || deleted != 2 {
		t.Fatalf("bulk delete = %d, %v; want 2, nil", deleted, err)
	}
}
