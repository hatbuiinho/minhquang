package event

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"time"
)

type Service struct {
	store    Store
	now      func() time.Time
	resolver RecipientResolver
}

func NewService(store Store, now func() time.Time) *Service {
	return &Service{store: store, now: now}
}

func (s *Service) SetRecipientResolver(resolver RecipientResolver) {
	s.resolver = resolver
}

func (s *Service) Create(ctx context.Context, input CreateInput) (Event, error) {
	if err := validateCreate(input); err != nil {
		return Event{}, err
	}

	now := s.now().UTC()
	eventID := newID()
	item := Event{
		ID:                 eventID,
		UserID:             strings.TrimSpace(input.UserID),
		Title:              strings.TrimSpace(input.Title),
		Description:        strings.TrimSpace(input.Description),
		StartsAt:           input.StartsAt.UTC(),
		Timezone:           normalizeTimezone(input.Timezone),
		Status:             StatusActive,
		AudienceType:       normalizeAudienceType(input.AudienceType),
		ReminderGeneration: 1,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	reminders, err := normalizeReminderRules(eventID, input.Reminders, now)
	if err != nil {
		return Event{}, err
	}
	item.Reminders = reminders

	recipients, err := s.resolveRecipients(ctx, item, RecipientInput{
		AudienceType: item.AudienceType,
		UserIDs:      input.RecipientUserIDs,
		GroupIDs:     input.RecipientGroupIDs,
	}, now)
	if err != nil {
		return Event{}, err
	}
	item.Recipients = recipients

	jobs := buildReminderJobs(item, recipients, now)
	return s.store.Create(ctx, item, recipients, jobs)
}

func (s *Service) List(ctx context.Context, userID string, options ListOptions) (EventPage, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return EventPage{}, fmt.Errorf("%w: user_id is required", ErrInvalidInput)
	}
	if options.Limit <= 0 {
		options.Limit = 20
	}
	if options.Limit > 100 {
		options.Limit = 100
	}

	cursor, err := decodeEventCursor(options.Cursor)
	if err != nil {
		return EventPage{}, err
	}

	return s.store.ListByUser(ctx, userID, options.Limit, cursor)
}

func nextEventCursor(items []Event, hasMore bool) string {
	if !hasMore || len(items) == 0 {
		return ""
	}

	last := items[len(items)-1]
	return base64.RawURLEncoding.EncodeToString([]byte(
		last.CreatedAt.UTC().Format(time.RFC3339Nano) + "|" + last.ID,
	))
}

func (s *Service) Get(ctx context.Context, userID string, id string) (Event, error) {
	userID = strings.TrimSpace(userID)
	id = strings.TrimSpace(id)
	if userID == "" || id == "" {
		return Event{}, fmt.Errorf("%w: user_id and id are required", ErrInvalidInput)
	}

	return s.store.GetByID(ctx, userID, id)
}

func (s *Service) ListUpcomingReminderJobs(ctx context.Context, userID string, limit int) ([]ReminderJob, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, fmt.Errorf("%w: user_id is required", ErrInvalidInput)
	}
	if limit <= 0 {
		limit = 50
	}

	return s.store.ListUpcomingReminderJobs(ctx, userID, s.now().UTC(), limit)
}

func (s *Service) ListReminderInbox(ctx context.Context, userID string, limit int) ([]ReminderJob, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, fmt.Errorf("%w: user_id is required", ErrInvalidInput)
	}
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	return s.store.ListReminderInbox(ctx, userID, s.now().UTC(), limit)
}

func (s *Service) MarkReminderJobRead(ctx context.Context, userID string, id string) (ReminderJob, error) {
	userID = strings.TrimSpace(userID)
	id = strings.TrimSpace(id)
	if userID == "" || id == "" {
		return ReminderJob{}, fmt.Errorf("%w: user_id and id are required", ErrInvalidInput)
	}

	return s.store.MarkReminderJobRead(ctx, userID, id, s.now().UTC())
}

func (s *Service) DismissReminderJob(ctx context.Context, userID string, id string) (ReminderJob, error) {
	userID = strings.TrimSpace(userID)
	id = strings.TrimSpace(id)
	if userID == "" || id == "" {
		return ReminderJob{}, fmt.Errorf("%w: user_id and id are required", ErrInvalidInput)
	}

	return s.store.DismissReminderJob(ctx, userID, id, s.now().UTC())
}

func (s *Service) SnoozeReminderJob(ctx context.Context, userID string, id string, delayMinutes int) (ReminderJob, error) {
	userID = strings.TrimSpace(userID)
	id = strings.TrimSpace(id)
	if userID == "" || id == "" {
		return ReminderJob{}, fmt.Errorf("%w: user_id and id are required", ErrInvalidInput)
	}
	if delayMinutes < 1 || delayMinutes > 7*24*60 {
		return ReminderJob{}, fmt.Errorf("%w: delay_minutes must be between 1 and 10080", ErrInvalidInput)
	}

	now := s.now().UTC()
	snoozed := ReminderJob{
		ID:            newReminderJobID(),
		UserID:        userID,
		Status:        ReminderJobPending,
		ScheduledAt:   now.Add(time.Duration(delayMinutes) * time.Minute),
		SnoozedFromID: id,
		SnoozedAt:     &now,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	return s.store.SnoozeReminderJob(ctx, userID, id, snoozed, now)
}

func (s *Service) Update(ctx context.Context, userID string, id string, input UpdateInput) (Event, error) {
	current, err := s.Get(ctx, userID, id)
	if err != nil {
		return Event{}, err
	}

	shouldRegenerateReminders := false
	shouldReplaceReminderJobs := false
	shouldReplaceRecipients := false
	recipientInput := RecipientInput{AudienceType: current.AudienceType}

	if input.Title != nil {
		title := strings.TrimSpace(*input.Title)
		if title == "" {
			return Event{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
		}
		current.Title = title
		shouldReplaceReminderJobs = true
	}

	if input.Description != nil {
		current.Description = strings.TrimSpace(*input.Description)
	}

	if input.StartsAt != nil {
		if input.StartsAt.IsZero() {
			return Event{}, fmt.Errorf("%w: starts_at is required", ErrInvalidInput)
		}
		current.StartsAt = input.StartsAt.UTC()
		shouldRegenerateReminders = true
		shouldReplaceReminderJobs = true
	}

	if input.Timezone != nil {
		current.Timezone = normalizeTimezone(*input.Timezone)
		shouldRegenerateReminders = true
		shouldReplaceReminderJobs = true
	}

	if input.Status != nil {
		if !validStatus(*input.Status) {
			return Event{}, ErrInvalidStatus
		}
		current.Status = *input.Status
		shouldReplaceReminderJobs = true
	}

	if input.AudienceType != nil {
		current.AudienceType = normalizeAudienceType(*input.AudienceType)
		recipientInput.AudienceType = current.AudienceType
		shouldReplaceRecipients = true
		shouldReplaceReminderJobs = true
	}
	if input.RecipientUserIDs != nil {
		recipientInput.UserIDs = *input.RecipientUserIDs
		shouldReplaceRecipients = true
		shouldReplaceReminderJobs = true
	}
	if input.RecipientGroupIDs != nil {
		recipientInput.GroupIDs = *input.RecipientGroupIDs
		shouldReplaceRecipients = true
		shouldReplaceReminderJobs = true
	}

	now := s.now().UTC()
	if input.Reminders != nil {
		reminders, err := normalizeReminderRules(current.ID, *input.Reminders, now)
		if err != nil {
			return Event{}, err
		}
		current.Reminders = reminders
		shouldRegenerateReminders = true
		shouldReplaceReminderJobs = true
	}

	if shouldRegenerateReminders {
		current.ReminderGeneration++
	}

	current.UpdatedAt = now
	recipients := current.Recipients
	if shouldReplaceRecipients {
		var err error
		recipients, err = s.resolveRecipients(ctx, current, recipientInput, now)
		if err != nil {
			return Event{}, err
		}
		current.Recipients = recipients
	}
	jobs := []ReminderJob(nil)
	if shouldReplaceReminderJobs {
		if len(recipients) == 0 {
			var err error
			recipients, err = s.store.ListRecipients(ctx, current.ID)
			if err != nil {
				return Event{}, err
			}
		}
		jobs = buildReminderJobs(current, recipients, now)
	}
	return s.store.Update(ctx, current, recipients, jobs, shouldReplaceRecipients, shouldReplaceReminderJobs)
}

func (s *Service) Delete(ctx context.Context, userID string, id string) error {
	userID = strings.TrimSpace(userID)
	id = strings.TrimSpace(id)
	if userID == "" || id == "" {
		return fmt.Errorf("%w: user_id and id are required", ErrInvalidInput)
	}

	return s.store.Delete(ctx, userID, id)
}

func validateCreate(input CreateInput) error {
	if strings.TrimSpace(input.UserID) == "" {
		return fmt.Errorf("%w: user_id is required", ErrInvalidInput)
	}
	if strings.TrimSpace(input.Title) == "" {
		return fmt.Errorf("%w: title is required", ErrInvalidInput)
	}
	if input.StartsAt.IsZero() {
		return fmt.Errorf("%w: starts_at is required", ErrInvalidInput)
	}

	if _, err := normalizeReminderRules("validation", input.Reminders, time.Now().UTC()); err != nil {
		return err
	}

	return nil
}

func normalizeTimezone(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "UTC"
	}
	return value
}

func normalizeAudienceType(value AudienceType) AudienceType {
	switch value {
	case AudienceSelectedUsers, AudienceSelectedGroups, AudienceAllUsers:
		return value
	default:
		return AudienceSelf
	}
}

func validStatus(status Status) bool {
	switch status {
	case StatusActive, StatusArchived, StatusCancelled:
		return true
	default:
		return false
	}
}

func normalizeReminderRules(eventID string, inputs []ReminderRuleInput, now time.Time) ([]ReminderRule, error) {
	rules := make([]ReminderRule, 0, len(inputs))
	seenOffsets := make(map[int]struct{}, len(inputs))

	for _, input := range inputs {
		if input.OffsetMinutes < 0 {
			return nil, fmt.Errorf("%w: offset_minutes must be greater than or equal to 0", ErrInvalidRule)
		}
		if _, ok := seenOffsets[input.OffsetMinutes]; ok {
			return nil, fmt.Errorf("%w: duplicate offset_minutes", ErrInvalidRule)
		}
		seenOffsets[input.OffsetMinutes] = struct{}{}

		enabled := true
		if input.Enabled != nil {
			enabled = *input.Enabled
		}

		channel := strings.TrimSpace(input.Channel)
		if channel == "" {
			channel = "push"
		}
		if channel != "push" {
			return nil, fmt.Errorf("%w: channel must be push", ErrInvalidRule)
		}

		importance := normalizeReminderImportance(input.Importance)
		if importance == "" {
			return nil, fmt.Errorf("%w: importance must be normal or urgent", ErrInvalidRule)
		}

		rules = append(rules, ReminderRule{
			ID:            newRuleID(),
			EventID:       eventID,
			OffsetMinutes: input.OffsetMinutes,
			Enabled:       enabled,
			Channel:       channel,
			Importance:    importance,
			CreatedAt:     now,
			UpdatedAt:     now,
		})
	}

	return rules, nil
}

func (s *Service) resolveRecipients(ctx context.Context, item Event, input RecipientInput, now time.Time) ([]EventRecipient, error) {
	input.AudienceType = normalizeAudienceType(input.AudienceType)
	if s.resolver == nil {
		input.AudienceType = AudienceSelf
	}

	resolved := []ResolvedRecipient{{
		UserID:     item.UserID,
		SourceType: RecipientSourceSelf,
	}}
	if s.resolver != nil {
		var err error
		resolved, err = s.resolver.ResolveEventRecipients(ctx, item.UserID, input)
		if err != nil {
			return nil, err
		}
	}

	recipients := make([]EventRecipient, 0, len(resolved))
	seen := make(map[string]struct{}, len(resolved))
	for _, recipient := range resolved {
		userID := strings.TrimSpace(recipient.UserID)
		if userID == "" {
			continue
		}
		if _, ok := seen[userID]; ok {
			continue
		}
		seen[userID] = struct{}{}
		recipients = append(recipients, EventRecipient{
			EventID:    item.ID,
			UserID:     userID,
			SourceType: recipient.SourceType,
			SourceID:   recipient.SourceID,
			CreatedAt:  now,
		})
	}
	if len(recipients) == 0 {
		return nil, fmt.Errorf("%w: recipients are required", ErrInvalidInput)
	}

	return recipients, nil
}

func buildReminderJobs(item Event, recipients []EventRecipient, now time.Time) []ReminderJob {
	if item.Status != StatusActive {
		return nil
	}

	jobs := make([]ReminderJob, 0, len(item.Reminders)*len(recipients))
	for _, rule := range item.Reminders {
		if !rule.Enabled {
			continue
		}

		scheduledAt := item.StartsAt.Add(-time.Duration(rule.OffsetMinutes) * time.Minute)
		if scheduledAt.Before(now) {
			continue
		}

		for _, recipient := range recipients {
			jobs = append(jobs, ReminderJob{
				ID:                 newReminderJobID(),
				UserID:             recipient.UserID,
				EventID:            item.ID,
				ReminderRuleID:     rule.ID,
				EventTitle:         item.Title,
				EventStartsAt:      item.StartsAt,
				OffsetMinutes:      rule.OffsetMinutes,
				Channel:            rule.Channel,
				Importance:         rule.Importance,
				Status:             ReminderJobPending,
				ScheduledAt:        scheduledAt,
				ReminderGeneration: item.ReminderGeneration,
				CreatedAt:          now,
				UpdatedAt:          now,
			})
		}
	}

	return jobs
}

func normalizeReminderImportance(value ReminderImportance) ReminderImportance {
	switch ReminderImportance(strings.TrimSpace(string(value))) {
	case "":
		return ReminderImportanceNormal
	case ReminderImportanceNormal:
		return ReminderImportanceNormal
	case ReminderImportanceUrgent:
		return ReminderImportanceUrgent
	default:
		return ""
	}
}

func decodeEventCursor(value string) (*EventCursor, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}

	decoded, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return nil, fmt.Errorf("%w: cursor is invalid", ErrInvalidInput)
	}

	parts := strings.SplitN(string(decoded), "|", 2)
	if len(parts) != 2 || strings.TrimSpace(parts[1]) == "" {
		return nil, fmt.Errorf("%w: cursor is invalid", ErrInvalidInput)
	}

	createdAt, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil {
		return nil, fmt.Errorf("%w: cursor is invalid", ErrInvalidInput)
	}

	return &EventCursor{CreatedAt: createdAt.UTC(), ID: parts[1]}, nil
}
