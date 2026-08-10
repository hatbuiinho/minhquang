package event

import (
	"context"
	"sort"
	"sync"
	"time"
)

type Store interface {
	Create(ctx context.Context, item Event, recipients []EventRecipient, jobs []ReminderJob) (Event, error)
	ListByUser(ctx context.Context, userID string, limit int, cursor *EventCursor) (EventPage, error)
	GetByID(ctx context.Context, userID string, id string) (Event, error)
	Update(ctx context.Context, item Event, recipients []EventRecipient, jobs []ReminderJob, replaceRecipients bool, replaceReminderJobs bool) (Event, error)
	Delete(ctx context.Context, userID string, id string) error
	ListUpcomingReminderJobs(ctx context.Context, userID string, now time.Time, limit int) ([]ReminderJob, error)
	ListReminderInbox(ctx context.Context, userID string, now time.Time, limit int) ([]ReminderJob, error)
	MarkReminderJobRead(ctx context.Context, userID string, id string, now time.Time) (ReminderJob, error)
	DismissReminderJob(ctx context.Context, userID string, id string, now time.Time) (ReminderJob, error)
	SnoozeReminderJob(ctx context.Context, userID string, id string, snoozed ReminderJob, now time.Time) (ReminderJob, error)
	ListRecipients(ctx context.Context, eventID string) ([]EventRecipient, error)
	ClaimDueReminderJobs(ctx context.Context, now time.Time, limit int) ([]ReminderJob, error)
	MarkReminderJobSent(ctx context.Context, id string, now time.Time) error
	MarkReminderJobFailed(ctx context.Context, id string, now time.Time) error
}

type MemoryStore struct {
	mu           sync.RWMutex
	events       map[string]Event
	userIDs      map[string][]string
	userPos      map[string]map[string]int
	eventKey     map[string]string
	reminderJobs map[string]ReminderJob
	eventJobIDs  map[string][]string
	recipients   map[string][]EventRecipient
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		events:       make(map[string]Event),
		userIDs:      make(map[string][]string),
		userPos:      make(map[string]map[string]int),
		eventKey:     make(map[string]string),
		reminderJobs: make(map[string]ReminderJob),
		eventJobIDs:  make(map[string][]string),
		recipients:   make(map[string][]EventRecipient),
	}
}

func (s *MemoryStore) Create(_ context.Context, item Event, recipients []EventRecipient, jobs []ReminderJob) (Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := userEventKey(item.UserID, item.ID)
	s.events[key] = item
	s.eventKey[item.ID] = key

	if s.userPos[item.UserID] == nil {
		s.userPos[item.UserID] = make(map[string]int)
	}
	s.userPos[item.UserID][item.ID] = len(s.userIDs[item.UserID])
	s.userIDs[item.UserID] = append(s.userIDs[item.UserID], item.ID)
	item.Recipients = recipients
	s.recipients[item.ID] = append([]EventRecipient(nil), recipients...)
	s.replaceReminderJobs(item.ID, jobs)

	return item, nil
}

func (s *MemoryStore) ListByUser(_ context.Context, userID string, limit int, cursor *EventCursor) (EventPage, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]Event, 0)
	for _, item := range s.events {
		for _, recipient := range s.recipients[item.ID] {
			if recipient.UserID == userID {
				item.Recipients = append([]EventRecipient(nil), s.recipients[item.ID]...)
				items = append(items, item)
				break
			}
		}
	}
	sortEventsByCreatedAtDesc(items)

	if cursor != nil {
		next := items[:0]
		for _, item := range items {
			if item.CreatedAt.Before(cursor.CreatedAt) ||
				(item.CreatedAt.Equal(cursor.CreatedAt) && item.ID < cursor.ID) {
				next = append(next, item)
			}
		}
		items = next
	}

	hasMore := len(items) > limit
	if hasMore {
		items = items[:limit]
	}

	return EventPage{
		Items:      items,
		NextCursor: nextEventCursor(items, hasMore),
		HasMore:    hasMore,
	}, nil
}

func (s *MemoryStore) GetByID(_ context.Context, userID string, id string) (Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, ok := s.events[userEventKey(userID, id)]
	if !ok {
		return Event{}, ErrNotFound
	}

	item.Recipients = append([]EventRecipient(nil), s.recipients[item.ID]...)
	return item, nil
}

func (s *MemoryStore) Update(_ context.Context, item Event, recipients []EventRecipient, jobs []ReminderJob, replaceRecipients bool, replaceReminderJobs bool) (Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := userEventKey(item.UserID, item.ID)
	if _, ok := s.events[key]; !ok {
		return Event{}, ErrNotFound
	}

	s.events[key] = item
	if replaceRecipients {
		s.recipients[item.ID] = append([]EventRecipient(nil), recipients...)
		item.Recipients = recipients
	}
	if replaceReminderJobs {
		s.replaceReminderJobs(item.ID, jobs)
	}
	return item, nil
}

func (s *MemoryStore) Delete(_ context.Context, userID string, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := userEventKey(userID, id)
	if _, ok := s.events[key]; !ok {
		return ErrNotFound
	}

	delete(s.events, key)
	delete(s.eventKey, id)
	delete(s.recipients, id)
	s.replaceReminderJobs(id, nil)

	pos, ok := s.userPos[userID][id]
	if ok {
		ids := s.userIDs[userID]
		last := len(ids) - 1
		lastID := ids[last]
		ids[pos] = lastID
		s.userPos[userID][lastID] = pos
		s.userIDs[userID] = ids[:last]
		delete(s.userPos[userID], id)
	}

	return nil
}

func (s *MemoryStore) ListUpcomingReminderJobs(_ context.Context, userID string, now time.Time, limit int) ([]ReminderJob, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit <= 0 {
		limit = 50
	}

	items := make([]ReminderJob, 0, limit)
	for _, job := range s.reminderJobs {
		if job.UserID != userID || job.Status != ReminderJobPending || job.DismissedAt != nil || job.ScheduledAt.Before(now) {
			continue
		}
		items = append(items, job)
	}
	sortReminderJobs(items)
	if len(items) > limit {
		items = items[:limit]
	}

	return items, nil
}

func (s *MemoryStore) ListReminderInbox(_ context.Context, userID string, now time.Time, limit int) ([]ReminderJob, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit <= 0 {
		limit = 50
	}

	items := make([]ReminderJob, 0, limit)
	for _, job := range s.reminderJobs {
		if job.UserID != userID || job.DismissedAt != nil || job.ScheduledAt.After(now) {
			continue
		}
		switch job.Status {
		case ReminderJobPending, ReminderJobSent, ReminderJobFailed:
			items = append(items, job)
		}
	}
	sortReminderJobsByScheduledAtDesc(items)
	if len(items) > limit {
		items = items[:limit]
	}

	return items, nil
}

func (s *MemoryStore) MarkReminderJobRead(_ context.Context, userID string, id string, now time.Time) (ReminderJob, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	job, ok := s.reminderJobs[id]
	if !ok || job.UserID != userID {
		return ReminderJob{}, ErrNotFound
	}
	if job.ReadAt == nil {
		job.ReadAt = &now
	}
	job.UpdatedAt = now
	s.reminderJobs[id] = job

	return job, nil
}

func (s *MemoryStore) DismissReminderJob(_ context.Context, userID string, id string, now time.Time) (ReminderJob, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	job, ok := s.reminderJobs[id]
	if !ok || job.UserID != userID {
		return ReminderJob{}, ErrNotFound
	}
	if job.DismissedAt == nil {
		job.DismissedAt = &now
	}
	job.UpdatedAt = now
	s.reminderJobs[id] = job

	return job, nil
}

func (s *MemoryStore) SnoozeReminderJob(_ context.Context, userID string, id string, snoozed ReminderJob, now time.Time) (ReminderJob, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	job, ok := s.reminderJobs[id]
	if !ok || job.UserID != userID {
		return ReminderJob{}, ErrNotFound
	}
	if job.DismissedAt == nil {
		job.DismissedAt = &now
	}
	job.UpdatedAt = now
	s.reminderJobs[id] = job

	snoozed.UserID = job.UserID
	snoozed.EventID = job.EventID
	snoozed.ReminderRuleID = job.ReminderRuleID
	snoozed.EventTitle = job.EventTitle
	snoozed.EventStartsAt = job.EventStartsAt
	snoozed.OffsetMinutes = job.OffsetMinutes
	snoozed.Channel = job.Channel
	snoozed.Importance = job.Importance
	snoozed.ReminderGeneration = job.ReminderGeneration
	s.reminderJobs[snoozed.ID] = snoozed
	s.eventJobIDs[snoozed.EventID] = append(s.eventJobIDs[snoozed.EventID], snoozed.ID)

	return snoozed, nil
}

func (s *MemoryStore) ListRecipients(_ context.Context, eventID string) ([]EventRecipient, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return append([]EventRecipient(nil), s.recipients[eventID]...), nil
}

func (s *MemoryStore) ClaimDueReminderJobs(_ context.Context, now time.Time, limit int) ([]ReminderJob, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if limit <= 0 {
		limit = 25
	}

	items := make([]ReminderJob, 0, limit)
	for id, job := range s.reminderJobs {
		if job.Status != ReminderJobPending || job.DismissedAt != nil || job.ScheduledAt.After(now) {
			continue
		}
		job.Status = ReminderJobProcessing
		job.UpdatedAt = now
		s.reminderJobs[id] = job
		items = append(items, job)
	}
	sortReminderJobs(items)
	if len(items) > limit {
		for _, job := range items[limit:] {
			job.Status = ReminderJobPending
			job.UpdatedAt = now
			s.reminderJobs[job.ID] = job
		}
		items = items[:limit]
	}

	return items, nil
}

func (s *MemoryStore) MarkReminderJobSent(_ context.Context, id string, now time.Time) error {
	return s.updateReminderJobStatus(id, ReminderJobSent, now, &now)
}

func (s *MemoryStore) MarkReminderJobFailed(_ context.Context, id string, now time.Time) error {
	return s.updateReminderJobStatus(id, ReminderJobFailed, now, nil)
}

func (s *MemoryStore) updateReminderJobStatus(id string, status ReminderJobStatus, now time.Time, sentAt *time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	job, ok := s.reminderJobs[id]
	if !ok {
		return ErrNotFound
	}
	job.Status = status
	job.SentAt = sentAt
	job.UpdatedAt = now
	s.reminderJobs[id] = job

	return nil
}

func (s *MemoryStore) replaceReminderJobs(eventID string, jobs []ReminderJob) {
	for _, id := range s.eventJobIDs[eventID] {
		delete(s.reminderJobs, id)
	}

	nextIDs := make([]string, len(jobs))
	for index, job := range jobs {
		s.reminderJobs[job.ID] = job
		nextIDs[index] = job.ID
	}
	s.eventJobIDs[eventID] = nextIDs
}

func sortReminderJobs(items []ReminderJob) {
	sort.Slice(items, func(left, right int) bool {
		if items[left].ScheduledAt.Equal(items[right].ScheduledAt) {
			return items[left].ID < items[right].ID
		}
		return items[left].ScheduledAt.Before(items[right].ScheduledAt)
	})
}

func sortReminderJobsByScheduledAtDesc(items []ReminderJob) {
	sort.Slice(items, func(left, right int) bool {
		if items[left].ScheduledAt.Equal(items[right].ScheduledAt) {
			return items[left].ID > items[right].ID
		}
		return items[left].ScheduledAt.After(items[right].ScheduledAt)
	})
}

func sortEventsByCreatedAtDesc(items []Event) {
	sort.Slice(items, func(left, right int) bool {
		if items[left].CreatedAt.Equal(items[right].CreatedAt) {
			return items[left].ID > items[right].ID
		}
		return items[left].CreatedAt.After(items[right].CreatedAt)
	})
}

func userEventKey(userID string, id string) string {
	return userID + ":" + id
}
