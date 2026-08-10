package event

import (
	"context"
	"testing"
	"time"
)

func TestServiceCreateListGetUpdateDelete(t *testing.T) {
	now := time.Date(2026, 8, 2, 10, 0, 0, 0, time.UTC)
	service := NewService(NewMemoryStore(), func() time.Time { return now })
	ctx := context.Background()

	startsAt := time.Date(2026, 9, 1, 9, 0, 0, 0, time.FixedZone("ICT", 7*60*60))
	created, err := service.Create(ctx, CreateInput{
		UserID:      "user_1",
		Title:       " Renew passport ",
		Description: " Prepare documents ",
		StartsAt:    startsAt,
		Timezone:    "Asia/Ho_Chi_Minh",
	})
	if err != nil {
		t.Fatalf("create event: %v", err)
	}
	if created.Title != "Renew passport" {
		t.Fatalf("expected trimmed title, got %q", created.Title)
	}
	if created.StartsAt.Location() != time.UTC {
		t.Fatalf("expected starts_at normalized to UTC")
	}

	page, err := service.List(ctx, "user_1", ListOptions{})
	if err != nil {
		t.Fatalf("list events: %v", err)
	}
	if len(page.Items) != 1 {
		t.Fatalf("expected one event, got %d", len(page.Items))
	}

	title := "Passport appointment"
	updated, err := service.Update(ctx, "user_1", created.ID, UpdateInput{Title: &title})
	if err != nil {
		t.Fatalf("update event: %v", err)
	}
	if updated.Title != title {
		t.Fatalf("expected updated title, got %q", updated.Title)
	}

	if err := service.Delete(ctx, "user_1", created.ID); err != nil {
		t.Fatalf("delete event: %v", err)
	}

	_, err = service.Get(ctx, "user_1", created.ID)
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestServiceRejectsInvalidStatus(t *testing.T) {
	service := NewService(NewMemoryStore(), time.Now)
	ctx := context.Background()

	created, err := service.Create(ctx, CreateInput{
		UserID:   "user_1",
		Title:    "Trip",
		StartsAt: time.Now().Add(time.Hour),
	})
	if err != nil {
		t.Fatalf("create event: %v", err)
	}

	status := Status("done")
	_, err = service.Update(ctx, "user_1", created.ID, UpdateInput{Status: &status})
	if err != ErrInvalidStatus {
		t.Fatalf("expected ErrInvalidStatus, got %v", err)
	}
}

func TestServiceReminderRulesAndGeneration(t *testing.T) {
	now := time.Date(2026, 8, 2, 10, 0, 0, 0, time.UTC)
	service := NewService(NewMemoryStore(), func() time.Time { return now })
	ctx := context.Background()

	created, err := service.Create(ctx, CreateInput{
		UserID:   "user_1",
		Title:    "Passport",
		StartsAt: now.Add(24 * time.Hour),
		Reminders: []ReminderRuleInput{
			{OffsetMinutes: 1440},
			{OffsetMinutes: 10080},
		},
	})
	if err != nil {
		t.Fatalf("create event: %v", err)
	}
	if created.ReminderGeneration != 1 {
		t.Fatalf("expected initial reminder generation 1, got %d", created.ReminderGeneration)
	}
	if len(created.Reminders) != 2 {
		t.Fatalf("expected two reminder rules, got %d", len(created.Reminders))
	}
	if !created.Reminders[0].Enabled || created.Reminders[0].Channel != "push" {
		t.Fatalf("expected default enabled push reminder, got %+v", created.Reminders[0])
	}
	if created.Reminders[0].Importance != ReminderImportanceNormal {
		t.Fatalf("expected default normal reminder importance, got %q", created.Reminders[0].Importance)
	}

	title := "Passport updated"
	updatedTitle, err := service.Update(ctx, "user_1", created.ID, UpdateInput{Title: &title})
	if err != nil {
		t.Fatalf("update title: %v", err)
	}
	if updatedTitle.ReminderGeneration != 1 {
		t.Fatalf("title-only update should not change reminder generation, got %d", updatedTitle.ReminderGeneration)
	}

	enabled := false
	reminders := []ReminderRuleInput{
		{OffsetMinutes: 60, Enabled: &enabled, Importance: ReminderImportanceUrgent},
	}
	updatedRules, err := service.Update(ctx, "user_1", created.ID, UpdateInput{Reminders: &reminders})
	if err != nil {
		t.Fatalf("update reminders: %v", err)
	}
	if updatedRules.ReminderGeneration != 2 {
		t.Fatalf("expected reminder generation 2, got %d", updatedRules.ReminderGeneration)
	}
	if len(updatedRules.Reminders) != 1 {
		t.Fatalf("expected one replacement reminder rule, got %d", len(updatedRules.Reminders))
	}
	if updatedRules.Reminders[0].Enabled {
		t.Fatal("expected disabled reminder rule")
	}
	if updatedRules.Reminders[0].Importance != ReminderImportanceUrgent {
		t.Fatalf("expected urgent reminder importance, got %q", updatedRules.Reminders[0].Importance)
	}
}

func TestServiceListsEventsByCreatedAtDescWithCursor(t *testing.T) {
	now := time.Date(2026, 8, 2, 10, 0, 0, 0, time.UTC)
	service := NewService(NewMemoryStore(), func() time.Time { return now })
	ctx := context.Background()

	first, err := service.Create(ctx, CreateInput{
		UserID:   "user_1",
		Title:    "First",
		StartsAt: now.Add(time.Hour),
	})
	if err != nil {
		t.Fatalf("create first event: %v", err)
	}

	now = now.Add(time.Minute)
	second, err := service.Create(ctx, CreateInput{
		UserID:   "user_1",
		Title:    "Second",
		StartsAt: now.Add(time.Hour),
	})
	if err != nil {
		t.Fatalf("create second event: %v", err)
	}

	firstPage, err := service.List(ctx, "user_1", ListOptions{Limit: 1})
	if err != nil {
		t.Fatalf("list first page: %v", err)
	}
	if len(firstPage.Items) != 1 || firstPage.Items[0].ID != second.ID {
		t.Fatalf("expected newest event first, got %+v", firstPage.Items)
	}
	if !firstPage.HasMore || firstPage.NextCursor == "" {
		t.Fatalf("expected next cursor, got %+v", firstPage)
	}

	secondPage, err := service.List(ctx, "user_1", ListOptions{Limit: 1, Cursor: firstPage.NextCursor})
	if err != nil {
		t.Fatalf("list second page: %v", err)
	}
	if len(secondPage.Items) != 1 || secondPage.Items[0].ID != first.ID {
		t.Fatalf("expected older event on second page, got %+v", secondPage.Items)
	}
	if secondPage.HasMore || secondPage.NextCursor != "" {
		t.Fatalf("expected final page, got %+v", secondPage)
	}
}

func TestServiceGeneratesUpcomingReminderJobs(t *testing.T) {
	now := time.Date(2026, 8, 2, 10, 0, 0, 0, time.UTC)
	service := NewService(NewMemoryStore(), func() time.Time { return now })
	ctx := context.Background()

	enabled := true
	disabled := false
	created, err := service.Create(ctx, CreateInput{
		UserID:   "user_1",
		Title:    "Passport",
		StartsAt: now.Add(48 * time.Hour),
		Reminders: []ReminderRuleInput{
			{OffsetMinutes: 60, Enabled: &enabled},
			{OffsetMinutes: 1440, Enabled: &enabled},
			{OffsetMinutes: 10080, Enabled: &enabled},
			{OffsetMinutes: 30, Enabled: &disabled},
		},
	})
	if err != nil {
		t.Fatalf("create event: %v", err)
	}

	jobs, err := service.ListUpcomingReminderJobs(ctx, "user_1", 10)
	if err != nil {
		t.Fatalf("list upcoming reminder jobs: %v", err)
	}
	if len(jobs) != 2 {
		t.Fatalf("expected two upcoming jobs, got %d", len(jobs))
	}
	if jobs[0].ScheduledAt.After(jobs[1].ScheduledAt) {
		t.Fatal("expected jobs sorted by scheduled_at")
	}
	if jobs[0].EventID != created.ID || jobs[0].EventTitle != created.Title {
		t.Fatalf("expected job to include event snapshot, got %+v", jobs[0])
	}
	if jobs[0].ReminderGeneration != created.ReminderGeneration {
		t.Fatalf("expected generation %d, got %d", created.ReminderGeneration, jobs[0].ReminderGeneration)
	}
}

func TestServiceReminderInboxReadAndDismiss(t *testing.T) {
	now := time.Date(2026, 8, 2, 10, 0, 0, 0, time.UTC)
	store := NewMemoryStore()
	service := NewService(store, func() time.Time { return now })
	ctx := context.Background()

	_, err := service.Create(ctx, CreateInput{
		UserID:   "user_1",
		Title:    "Passport",
		StartsAt: now.Add(3 * time.Hour),
		Reminders: []ReminderRuleInput{
			{OffsetMinutes: 0},
			{OffsetMinutes: 90},
		},
	})
	if err != nil {
		t.Fatalf("create event: %v", err)
	}

	inbox, err := service.ListReminderInbox(ctx, "user_1", 10)
	if err != nil {
		t.Fatalf("list reminder inbox before due: %v", err)
	}
	if len(inbox) != 0 {
		t.Fatalf("expected no due reminders before scheduled time, got %d", len(inbox))
	}

	now = now.Add(3 * time.Hour)
	inbox, err = service.ListReminderInbox(ctx, "user_1", 10)
	if err != nil {
		t.Fatalf("list reminder inbox: %v", err)
	}
	if len(inbox) != 2 {
		t.Fatalf("expected two due reminders, got %d", len(inbox))
	}

	read, err := service.MarkReminderJobRead(ctx, "user_1", inbox[0].ID)
	if err != nil {
		t.Fatalf("mark reminder read: %v", err)
	}
	if read.ReadAt == nil {
		t.Fatal("expected read_at to be set")
	}

	dismissed, err := service.DismissReminderJob(ctx, "user_1", inbox[0].ID)
	if err != nil {
		t.Fatalf("dismiss reminder: %v", err)
	}
	if dismissed.DismissedAt == nil {
		t.Fatal("expected dismissed_at to be set")
	}

	inbox, err = service.ListReminderInbox(ctx, "user_1", 10)
	if err != nil {
		t.Fatalf("list reminder inbox after dismiss: %v", err)
	}
	if len(inbox) != 1 {
		t.Fatalf("expected one reminder after dismiss, got %d", len(inbox))
	}

	if _, err := service.MarkReminderJobRead(ctx, "other_user", inbox[0].ID); err != ErrNotFound {
		t.Fatalf("expected ErrNotFound for another user, got %v", err)
	}
}

func TestServiceSnoozesReminderJob(t *testing.T) {
	now := time.Date(2026, 8, 2, 10, 0, 0, 0, time.UTC)
	service := NewService(NewMemoryStore(), func() time.Time { return now })
	ctx := context.Background()

	_, err := service.Create(ctx, CreateInput{
		UserID:   "user_1",
		Title:    "Passport",
		StartsAt: now.Add(time.Hour),
		Reminders: []ReminderRuleInput{
			{OffsetMinutes: 0},
		},
	})
	if err != nil {
		t.Fatalf("create event: %v", err)
	}

	now = now.Add(time.Hour)
	inbox, err := service.ListReminderInbox(ctx, "user_1", 10)
	if err != nil {
		t.Fatalf("list reminder inbox: %v", err)
	}
	if len(inbox) != 1 {
		t.Fatalf("expected one due reminder, got %d", len(inbox))
	}

	snoozed, err := service.SnoozeReminderJob(ctx, "user_1", inbox[0].ID, 15)
	if err != nil {
		t.Fatalf("snooze reminder: %v", err)
	}
	if snoozed.SnoozedFromID != inbox[0].ID || snoozed.SnoozedAt == nil {
		t.Fatalf("expected snooze metadata, got %+v", snoozed)
	}
	if !snoozed.ScheduledAt.Equal(now.Add(15 * time.Minute)) {
		t.Fatalf("expected snoozed scheduled_at after delay, got %s", snoozed.ScheduledAt)
	}
	if snoozed.EventTitle != inbox[0].EventTitle || snoozed.EventID != inbox[0].EventID {
		t.Fatalf("expected snoozed job to keep event snapshot, got %+v", snoozed)
	}

	inbox, err = service.ListReminderInbox(ctx, "user_1", 10)
	if err != nil {
		t.Fatalf("list reminder inbox after snooze: %v", err)
	}
	if len(inbox) != 0 {
		t.Fatalf("expected inbox empty after snooze, got %d", len(inbox))
	}

	upcoming, err := service.ListUpcomingReminderJobs(ctx, "user_1", 10)
	if err != nil {
		t.Fatalf("list upcoming reminders: %v", err)
	}
	if len(upcoming) != 1 || upcoming[0].ID != snoozed.ID {
		t.Fatalf("expected snoozed reminder in upcoming, got %+v", upcoming)
	}
}

func TestServiceDismissesUpcomingReminderBeforeDue(t *testing.T) {
	now := time.Date(2026, 8, 2, 10, 0, 0, 0, time.UTC)
	store := NewMemoryStore()
	service := NewService(store, func() time.Time { return now })
	ctx := context.Background()

	_, err := service.Create(ctx, CreateInput{
		UserID:   "user_1",
		Title:    "Visa renewal",
		StartsAt: now.Add(24 * time.Hour),
		Reminders: []ReminderRuleInput{
			{OffsetMinutes: 60},
		},
	})
	if err != nil {
		t.Fatalf("create event: %v", err)
	}

	upcoming, err := service.ListUpcomingReminderJobs(ctx, "user_1", 10)
	if err != nil {
		t.Fatalf("list upcoming reminders: %v", err)
	}
	if len(upcoming) != 1 {
		t.Fatalf("expected one upcoming reminder, got %d", len(upcoming))
	}

	dismissed, err := service.DismissReminderJob(ctx, "user_1", upcoming[0].ID)
	if err != nil {
		t.Fatalf("dismiss upcoming reminder: %v", err)
	}
	if dismissed.DismissedAt == nil {
		t.Fatal("expected dismissed_at to be set")
	}

	upcoming, err = service.ListUpcomingReminderJobs(ctx, "user_1", 10)
	if err != nil {
		t.Fatalf("list upcoming reminders after dismiss: %v", err)
	}
	if len(upcoming) != 0 {
		t.Fatalf("expected no upcoming reminders after dismiss, got %d", len(upcoming))
	}

	claimed, err := store.ClaimDueReminderJobs(ctx, now.Add(24*time.Hour), 10)
	if err != nil {
		t.Fatalf("claim due reminders: %v", err)
	}
	if len(claimed) != 0 {
		t.Fatalf("expected dismissed reminder not to be claimed, got %d", len(claimed))
	}
}

func TestServiceGeneratesReminderJobsForDedupedRecipients(t *testing.T) {
	now := time.Date(2026, 8, 2, 10, 0, 0, 0, time.UTC)
	store := NewMemoryStore()
	service := NewService(store, func() time.Time { return now })
	service.SetRecipientResolver(staticRecipientResolver{
		recipients: []ResolvedRecipient{
			{UserID: "user_2", SourceType: RecipientSourceGroup, SourceID: "group_1"},
			{UserID: "user_2", SourceType: RecipientSourceGroup, SourceID: "group_2"},
		},
	})
	ctx := context.Background()

	created, err := service.Create(ctx, CreateInput{
		UserID:            "owner",
		Title:             "Meeting",
		StartsAt:          now.Add(time.Hour),
		AudienceType:      AudienceSelectedGroups,
		RecipientGroupIDs: []string{"group_1", "group_2"},
		Reminders:         []ReminderRuleInput{{OffsetMinutes: 0}},
	})
	if err != nil {
		t.Fatalf("create event: %v", err)
	}
	if len(created.Recipients) != 1 {
		t.Fatalf("expected deduped recipient, got %d", len(created.Recipients))
	}

	jobs, err := store.ListUpcomingReminderJobs(ctx, "user_2", now, 10)
	if err != nil {
		t.Fatalf("list upcoming jobs: %v", err)
	}
	if len(jobs) != 1 {
		t.Fatalf("expected one reminder job, got %d", len(jobs))
	}
}

func TestServiceReplacesReminderJobsWhenEventChanges(t *testing.T) {
	now := time.Date(2026, 8, 2, 10, 0, 0, 0, time.UTC)
	service := NewService(NewMemoryStore(), func() time.Time { return now })
	ctx := context.Background()

	created, err := service.Create(ctx, CreateInput{
		UserID:   "user_1",
		Title:    "Passport",
		StartsAt: now.Add(48 * time.Hour),
		Reminders: []ReminderRuleInput{
			{OffsetMinutes: 60},
		},
	})
	if err != nil {
		t.Fatalf("create event: %v", err)
	}

	title := "Passport updated"
	updated, err := service.Update(ctx, "user_1", created.ID, UpdateInput{Title: &title})
	if err != nil {
		t.Fatalf("update title: %v", err)
	}
	if updated.ReminderGeneration != created.ReminderGeneration {
		t.Fatalf("title-only update should keep generation, got %d", updated.ReminderGeneration)
	}

	jobs, err := service.ListUpcomingReminderJobs(ctx, "user_1", 10)
	if err != nil {
		t.Fatalf("list upcoming reminder jobs: %v", err)
	}
	if len(jobs) != 1 {
		t.Fatalf("expected one upcoming job, got %d", len(jobs))
	}
	if jobs[0].EventTitle != title {
		t.Fatalf("expected updated event title in job, got %q", jobs[0].EventTitle)
	}

	status := StatusCancelled
	if _, err := service.Update(ctx, "user_1", created.ID, UpdateInput{Status: &status}); err != nil {
		t.Fatalf("cancel event: %v", err)
	}
	jobs, err = service.ListUpcomingReminderJobs(ctx, "user_1", 10)
	if err != nil {
		t.Fatalf("list upcoming reminder jobs after cancel: %v", err)
	}
	if len(jobs) != 0 {
		t.Fatalf("expected no upcoming jobs after cancel, got %d", len(jobs))
	}
}

func TestServiceRejectsInvalidReminderRule(t *testing.T) {
	service := NewService(NewMemoryStore(), time.Now)
	ctx := context.Background()

	_, err := service.Create(ctx, CreateInput{
		UserID:   "user_1",
		Title:    "Trip",
		StartsAt: time.Now().Add(time.Hour),
		Reminders: []ReminderRuleInput{
			{OffsetMinutes: -1},
		},
	})
	if err == nil {
		t.Fatal("expected invalid reminder rule error")
	}
}

type staticRecipientResolver struct {
	recipients []ResolvedRecipient
}

func (r staticRecipientResolver) ResolveEventRecipients(_ context.Context, _ string, _ RecipientInput) ([]ResolvedRecipient, error) {
	return r.recipients, nil
}
