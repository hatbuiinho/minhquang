package reminder

import (
	"context"
	"testing"
	"time"

	"reminder/be/internal/device"
	"reminder/be/internal/event"
	"reminder/be/internal/push"
)

type recordingSender struct {
	messages []push.Message
}

func (s *recordingSender) Send(_ context.Context, message push.Message) error {
	s.messages = append(s.messages, message)
	return nil
}

func TestWorkerProcessesDueReminderJob(t *testing.T) {
	now := time.Date(2026, 8, 2, 10, 0, 0, 0, time.UTC)
	eventStore := event.NewMemoryStore()
	eventService := event.NewService(eventStore, func() time.Time { return now })
	deviceService := device.NewService(device.NewMemoryStore(), func() time.Time { return now })
	sender := &recordingSender{}
	ctx := context.Background()

	_, err := deviceService.Register(ctx, device.RegisterInput{
		UserID:    "user_1",
		Platform:  device.PlatformAndroid,
		PushToken: "token_1",
	})
	if err != nil {
		t.Fatalf("register device: %v", err)
	}

	_, err = eventService.Create(ctx, event.CreateInput{
		UserID:   "user_1",
		Title:    "Passport",
		StartsAt: now,
		Reminders: []event.ReminderRuleInput{
			{OffsetMinutes: 0, Importance: event.ReminderImportanceUrgent},
		},
	})
	if err != nil {
		t.Fatalf("create event: %v", err)
	}

	worker := NewWorker(eventStore, deviceService, sender, func() time.Time { return now }, Config{})
	if err := worker.ProcessDue(ctx); err != nil {
		t.Fatalf("process due jobs: %v", err)
	}

	if len(sender.messages) != 1 {
		t.Fatalf("expected one push message, got %d", len(sender.messages))
	}
	if sender.messages[0].Token != "token_1" {
		t.Fatalf("expected token_1, got %q", sender.messages[0].Token)
	}
	if sender.messages[0].Importance != push.ImportanceUrgent {
		t.Fatalf("expected urgent push importance, got %q", sender.messages[0].Importance)
	}
}
