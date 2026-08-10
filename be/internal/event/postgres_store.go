package event

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStore struct {
	pool *pgxpool.Pool
}

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{pool: pool}
}

func (s *PostgresStore) Create(ctx context.Context, item Event, recipients []EventRecipient, jobs []ReminderJob) (Event, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Event{}, fmt.Errorf("begin create event: %w", err)
	}
	defer rollback(ctx, tx)

	const query = `
			INSERT INTO events (
				id, user_id, title, description, starts_at, timezone, status, audience_type, reminder_generation, created_at, updated_at
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		`

	_, err = tx.Exec(ctx, query,
		item.ID,
		item.UserID,
		item.Title,
		item.Description,
		item.StartsAt,
		item.Timezone,
		item.Status,
		item.AudienceType,
		item.ReminderGeneration,
		item.CreatedAt,
		item.UpdatedAt,
	)
	if err != nil {
		return Event{}, fmt.Errorf("insert event: %w", err)
	}
	if err := insertReminderRules(ctx, tx, item.Reminders); err != nil {
		return Event{}, err
	}
	if err := insertEventRecipients(ctx, tx, recipients); err != nil {
		return Event{}, err
	}
	if err := insertReminderJobs(ctx, tx, jobs); err != nil {
		return Event{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Event{}, fmt.Errorf("commit create event: %w", err)
	}

	return item, nil
}

func (s *PostgresStore) ListByUser(ctx context.Context, userID string, limit int, cursor *EventCursor) (EventPage, error) {
	const firstPageQuery = `
			SELECT id, user_id, title, description, starts_at, timezone, status, audience_type, reminder_generation, created_at, updated_at
			FROM events
			WHERE EXISTS (
				SELECT 1 FROM event_recipients
				WHERE event_recipients.event_id = events.id AND event_recipients.user_id = $1
			)
			ORDER BY created_at DESC, id DESC
			LIMIT $2
	`
	const nextPageQuery = `
			SELECT id, user_id, title, description, starts_at, timezone, status, audience_type, reminder_generation, created_at, updated_at
			FROM events
			WHERE EXISTS (
				SELECT 1 FROM event_recipients
				WHERE event_recipients.event_id = events.id AND event_recipients.user_id = $1
			)
			AND (created_at, id) < ($2, $3)
			ORDER BY created_at DESC, id DESC
			LIMIT $4
	`

	queryLimit := limit + 1
	var rows pgx.Rows
	var err error
	if cursor == nil {
		rows, err = s.pool.Query(ctx, firstPageQuery, userID, queryLimit)
	} else {
		rows, err = s.pool.Query(ctx, nextPageQuery, userID, cursor.CreatedAt, cursor.ID, queryLimit)
	}

	if err != nil {
		return EventPage{}, fmt.Errorf("list events: %w", err)
	}
	defer rows.Close()

	items := make([]Event, 0, queryLimit)
	for rows.Next() {
		item, err := scanEvent(rows)
		if err != nil {
			return EventPage{}, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return EventPage{}, fmt.Errorf("iterate events: %w", err)
	}

	hasMore := len(items) > limit
	if hasMore {
		items = items[:limit]
	}
	if err := s.attachReminderRules(ctx, items); err != nil {
		return EventPage{}, err
	}

	return EventPage{
		Items:      items,
		NextCursor: nextEventCursor(items, hasMore),
		HasMore:    hasMore,
	}, nil
}

func (s *PostgresStore) GetByID(ctx context.Context, userID string, id string) (Event, error) {
	const query = `
			SELECT id, user_id, title, description, starts_at, timezone, status, audience_type, reminder_generation, created_at, updated_at
			FROM events
			WHERE id = $2 AND EXISTS (
				SELECT 1 FROM event_recipients
				WHERE event_recipients.event_id = events.id AND event_recipients.user_id = $1
			)
	`

	item, err := scanEvent(s.pool.QueryRow(ctx, query, userID, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Event{}, ErrNotFound
		}
		return Event{}, err
	}
	rules, err := s.listReminderRules(ctx, item.ID)
	if err != nil {
		return Event{}, err
	}
	item.Reminders = rules
	recipients, err := s.ListRecipients(ctx, item.ID)
	if err != nil {
		return Event{}, err
	}
	item.Recipients = recipients

	return item, nil
}

func (s *PostgresStore) Update(ctx context.Context, item Event, recipients []EventRecipient, jobs []ReminderJob, replaceRecipients bool, replaceReminderJobs bool) (Event, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Event{}, fmt.Errorf("begin update event: %w", err)
	}
	defer rollback(ctx, tx)

	const query = `
		UPDATE events
		SET title = $3,
			description = $4,
			starts_at = $5,
				timezone = $6,
				status = $7,
				audience_type = $8,
				reminder_generation = $9,
				updated_at = $10
			WHERE user_id = $1 AND id = $2
			RETURNING id, user_id, title, description, starts_at, timezone, status, audience_type, reminder_generation, created_at, updated_at
	`

	updated, err := scanEvent(tx.QueryRow(ctx, query,
		item.UserID,
		item.ID,
		item.Title,
		item.Description,
		item.StartsAt,
		item.Timezone,
		item.Status,
		item.AudienceType,
		item.ReminderGeneration,
		item.UpdatedAt,
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Event{}, ErrNotFound
		}
		return Event{}, err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM reminder_rules WHERE event_id = $1`, item.ID); err != nil {
		return Event{}, fmt.Errorf("delete reminder rules: %w", err)
	}
	if err := insertReminderRules(ctx, tx, item.Reminders); err != nil {
		return Event{}, err
	}
	if replaceRecipients {
		if _, err := tx.Exec(ctx, `DELETE FROM event_recipients WHERE event_id = $1`, item.ID); err != nil {
			return Event{}, fmt.Errorf("delete event recipients: %w", err)
		}
		if err := insertEventRecipients(ctx, tx, recipients); err != nil {
			return Event{}, err
		}
		updated.Recipients = recipients
	}
	if !replaceRecipients {
		updated.Recipients = recipients
	}
	if replaceReminderJobs {
		if _, err := tx.Exec(ctx, `DELETE FROM reminder_jobs WHERE event_id = $1 AND status = $2`, item.ID, ReminderJobPending); err != nil {
			return Event{}, fmt.Errorf("delete reminder jobs: %w", err)
		}
		if err := insertReminderJobs(ctx, tx, jobs); err != nil {
			return Event{}, err
		}
	}
	updated.Reminders = item.Reminders
	if err := tx.Commit(ctx); err != nil {
		return Event{}, fmt.Errorf("commit update event: %w", err)
	}

	return updated, nil
}

func (s *PostgresStore) Delete(ctx context.Context, userID string, id string) error {
	const query = `
		DELETE FROM events
		WHERE user_id = $1 AND id = $2
	`

	tag, err := s.pool.Exec(ctx, query, userID, id)
	if err != nil {
		return fmt.Errorf("delete event: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (s *PostgresStore) ListUpcomingReminderJobs(ctx context.Context, userID string, now time.Time, limit int) ([]ReminderJob, error) {
	const query = `
		SELECT id, user_id, event_id, reminder_rule_id, event_title, event_starts_at,
			offset_minutes, channel, importance, status, scheduled_at, sent_at, read_at, dismissed_at,
			COALESCE(snoozed_from_id, ''), snoozed_at, cancelled_at,
			reminder_generation, created_at, updated_at
		FROM reminder_jobs
		WHERE user_id = $1 AND status = $2 AND dismissed_at IS NULL AND scheduled_at >= $3
		ORDER BY scheduled_at ASC, id ASC
		LIMIT $4
	`

	rows, err := s.pool.Query(ctx, query, userID, ReminderJobPending, now, limit)
	if err != nil {
		return nil, fmt.Errorf("list upcoming reminder jobs: %w", err)
	}
	defer rows.Close()

	items := make([]ReminderJob, 0, limit)
	for rows.Next() {
		item, err := scanReminderJob(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate upcoming reminder jobs: %w", err)
	}

	return items, nil
}

func (s *PostgresStore) ListReminderInbox(ctx context.Context, userID string, now time.Time, limit int) ([]ReminderJob, error) {
	const query = `
		SELECT id, user_id, event_id, reminder_rule_id, event_title, event_starts_at,
			offset_minutes, channel, importance, status, scheduled_at, sent_at, read_at, dismissed_at,
			COALESCE(snoozed_from_id, ''), snoozed_at, cancelled_at,
			reminder_generation, created_at, updated_at
		FROM reminder_jobs
		WHERE user_id = $1
			AND dismissed_at IS NULL
			AND scheduled_at <= $2
			AND status = ANY($3)
		ORDER BY scheduled_at DESC, id DESC
		LIMIT $4
	`

	rows, err := s.pool.Query(ctx, query, userID, now, []ReminderJobStatus{
		ReminderJobPending,
		ReminderJobSent,
		ReminderJobFailed,
	}, limit)
	if err != nil {
		return nil, fmt.Errorf("list reminder inbox: %w", err)
	}
	defer rows.Close()

	items := make([]ReminderJob, 0, limit)
	for rows.Next() {
		item, err := scanReminderJob(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate reminder inbox: %w", err)
	}

	return items, nil
}

func (s *PostgresStore) MarkReminderJobRead(ctx context.Context, userID string, id string, now time.Time) (ReminderJob, error) {
	const query = `
		UPDATE reminder_jobs
		SET read_at = COALESCE(read_at, $3), updated_at = $3
		WHERE user_id = $1 AND id = $2
		RETURNING id, user_id, event_id, reminder_rule_id, event_title, event_starts_at,
			offset_minutes, channel, importance, status, scheduled_at, sent_at, read_at, dismissed_at,
			COALESCE(snoozed_from_id, ''), snoozed_at, cancelled_at,
			reminder_generation, created_at, updated_at
	`

	item, err := scanReminderJob(s.pool.QueryRow(ctx, query, userID, id, now))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ReminderJob{}, ErrNotFound
		}
		return ReminderJob{}, fmt.Errorf("mark reminder job read: %w", err)
	}

	return item, nil
}

func (s *PostgresStore) DismissReminderJob(ctx context.Context, userID string, id string, now time.Time) (ReminderJob, error) {
	const query = `
		UPDATE reminder_jobs
		SET dismissed_at = COALESCE(dismissed_at, $3), updated_at = $3
		WHERE user_id = $1 AND id = $2
		RETURNING id, user_id, event_id, reminder_rule_id, event_title, event_starts_at,
			offset_minutes, channel, importance, status, scheduled_at, sent_at, read_at, dismissed_at,
			COALESCE(snoozed_from_id, ''), snoozed_at, cancelled_at,
			reminder_generation, created_at, updated_at
	`

	item, err := scanReminderJob(s.pool.QueryRow(ctx, query, userID, id, now))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ReminderJob{}, ErrNotFound
		}
		return ReminderJob{}, fmt.Errorf("dismiss reminder job: %w", err)
	}

	return item, nil
}

func (s *PostgresStore) SnoozeReminderJob(ctx context.Context, userID string, id string, snoozed ReminderJob, now time.Time) (ReminderJob, error) {
	const query = `
		WITH source_job AS (
			UPDATE reminder_jobs
			SET dismissed_at = COALESCE(dismissed_at, $4), updated_at = $4
			WHERE user_id = $1 AND id = $2
			RETURNING *
		)
		INSERT INTO reminder_jobs (
			id, user_id, event_id, reminder_rule_id, event_title, event_starts_at,
			offset_minutes, channel, importance, status, scheduled_at, sent_at, read_at, dismissed_at,
			snoozed_from_id, snoozed_at, cancelled_at, reminder_generation, created_at, updated_at
		)
		SELECT
			$3, user_id, event_id, reminder_rule_id, event_title, event_starts_at,
			offset_minutes, channel, importance, $5, $6, NULL, NULL, NULL,
			id, $4, NULL, reminder_generation, $4, $4
		FROM source_job
		RETURNING id, user_id, event_id, reminder_rule_id, event_title, event_starts_at,
			offset_minutes, channel, importance, status, scheduled_at, sent_at, read_at, dismissed_at,
			COALESCE(snoozed_from_id, ''), snoozed_at, cancelled_at,
			reminder_generation, created_at, updated_at
	`

	item, err := scanReminderJob(s.pool.QueryRow(
		ctx,
		query,
		userID,
		id,
		snoozed.ID,
		now,
		ReminderJobPending,
		snoozed.ScheduledAt,
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ReminderJob{}, ErrNotFound
		}
		return ReminderJob{}, fmt.Errorf("snooze reminder job: %w", err)
	}

	return item, nil
}

func (s *PostgresStore) ClaimDueReminderJobs(ctx context.Context, now time.Time, limit int) ([]ReminderJob, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin claim reminder jobs: %w", err)
	}
	defer rollback(ctx, tx)

	const query = `
		WITH due_jobs AS (
			SELECT id
			FROM reminder_jobs
			WHERE status = $1 AND dismissed_at IS NULL AND scheduled_at <= $2
			ORDER BY scheduled_at ASC, id ASC
			LIMIT $3
			FOR UPDATE SKIP LOCKED
		)
		UPDATE reminder_jobs
		SET status = $4, updated_at = $2
		WHERE id IN (SELECT id FROM due_jobs)
		RETURNING id, user_id, event_id, reminder_rule_id, event_title, event_starts_at,
			offset_minutes, channel, importance, status, scheduled_at, sent_at, read_at, dismissed_at,
			COALESCE(snoozed_from_id, ''), snoozed_at, cancelled_at,
			reminder_generation, created_at, updated_at
	`

	rows, err := tx.Query(ctx, query, ReminderJobPending, now, limit, ReminderJobProcessing)
	if err != nil {
		return nil, fmt.Errorf("claim due reminder jobs: %w", err)
	}
	defer rows.Close()

	items := make([]ReminderJob, 0, limit)
	for rows.Next() {
		item, err := scanReminderJob(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate claimed reminder jobs: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit claim reminder jobs: %w", err)
	}

	return items, nil
}

func (s *PostgresStore) MarkReminderJobSent(ctx context.Context, id string, now time.Time) error {
	const query = `
		UPDATE reminder_jobs
		SET status = $2, sent_at = $3, updated_at = $3
		WHERE id = $1
	`

	tag, err := s.pool.Exec(ctx, query, id, ReminderJobSent, now)
	if err != nil {
		return fmt.Errorf("mark reminder job sent: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (s *PostgresStore) MarkReminderJobFailed(ctx context.Context, id string, now time.Time) error {
	const query = `
		UPDATE reminder_jobs
		SET status = $2, updated_at = $3
		WHERE id = $1
	`

	tag, err := s.pool.Exec(ctx, query, id, ReminderJobFailed, now)
	if err != nil {
		return fmt.Errorf("mark reminder job failed: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

type eventScanner interface {
	Scan(dest ...any) error
}

func scanEvent(scanner eventScanner) (Event, error) {
	var item Event
	if err := scanner.Scan(
		&item.ID,
		&item.UserID,
		&item.Title,
		&item.Description,
		&item.StartsAt,
		&item.Timezone,
		&item.Status,
		&item.AudienceType,
		&item.ReminderGeneration,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		return Event{}, fmt.Errorf("scan event: %w", err)
	}

	return item, nil
}

type txExecutor interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

func insertReminderRules(ctx context.Context, tx txExecutor, rules []ReminderRule) error {
	const query = `
		INSERT INTO reminder_rules (
			id, event_id, offset_minutes, enabled, channel, importance, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	for _, rule := range rules {
		_, err := tx.Exec(ctx, query,
			rule.ID,
			rule.EventID,
			rule.OffsetMinutes,
			rule.Enabled,
			rule.Channel,
			rule.Importance,
			rule.CreatedAt,
			rule.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("insert reminder rule: %w", err)
		}
	}

	return nil
}

func insertReminderJobs(ctx context.Context, tx txExecutor, jobs []ReminderJob) error {
	const query = `
		INSERT INTO reminder_jobs (
			id, user_id, event_id, reminder_rule_id, event_title, event_starts_at,
			offset_minutes, channel, importance, status, scheduled_at, sent_at, read_at, dismissed_at,
			snoozed_from_id, snoozed_at, cancelled_at,
			reminder_generation, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, NULLIF($15, ''), $16, $17, $18, $19, $20)
	`

	for _, job := range jobs {
		_, err := tx.Exec(ctx, query,
			job.ID,
			job.UserID,
			job.EventID,
			job.ReminderRuleID,
			job.EventTitle,
			job.EventStartsAt,
			job.OffsetMinutes,
			job.Channel,
			job.Importance,
			job.Status,
			job.ScheduledAt,
			job.SentAt,
			job.ReadAt,
			job.DismissedAt,
			job.SnoozedFromID,
			job.SnoozedAt,
			job.CancelledAt,
			job.ReminderGeneration,
			job.CreatedAt,
			job.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("insert reminder job: %w", err)
		}
	}

	return nil
}

func insertEventRecipients(ctx context.Context, tx txExecutor, recipients []EventRecipient) error {
	const query = `
		INSERT INTO event_recipients (
			event_id, user_id, source_type, source_id, created_at
		)
		VALUES ($1, $2, $3, NULLIF($4, ''), $5)
		ON CONFLICT (event_id, user_id) DO NOTHING
	`

	for _, recipient := range recipients {
		_, err := tx.Exec(ctx, query,
			recipient.EventID,
			recipient.UserID,
			recipient.SourceType,
			recipient.SourceID,
			recipient.CreatedAt,
		)
		if err != nil {
			return fmt.Errorf("insert event recipient: %w", err)
		}
	}

	return nil
}

func (s *PostgresStore) ListRecipients(ctx context.Context, eventID string) ([]EventRecipient, error) {
	const query = `
		SELECT event_id, user_id, source_type, COALESCE(source_id, ''), created_at
		FROM event_recipients
		WHERE event_id = $1
		ORDER BY user_id ASC
	`

	rows, err := s.pool.Query(ctx, query, eventID)
	if err != nil {
		return nil, fmt.Errorf("list event recipients: %w", err)
	}
	defer rows.Close()

	items := make([]EventRecipient, 0)
	for rows.Next() {
		item, err := scanEventRecipient(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate event recipients: %w", err)
	}

	return items, nil
}

func (s *PostgresStore) attachReminderRules(ctx context.Context, items []Event) error {
	if len(items) == 0 {
		return nil
	}

	eventIDs := make([]string, len(items))
	eventIndexes := make(map[string]int, len(items))
	for index, item := range items {
		eventIDs[index] = item.ID
		eventIndexes[item.ID] = index
	}

	const query = `
		SELECT id, event_id, offset_minutes, enabled, channel, importance, created_at, updated_at
		FROM reminder_rules
		WHERE event_id = ANY($1)
		ORDER BY event_id ASC, offset_minutes ASC, id ASC
	`

	rows, err := s.pool.Query(ctx, query, eventIDs)
	if err != nil {
		return fmt.Errorf("list reminder rules: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		rule, err := scanReminderRule(rows)
		if err != nil {
			return err
		}
		index, ok := eventIndexes[rule.EventID]
		if ok {
			items[index].Reminders = append(items[index].Reminders, rule)
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate reminder rules: %w", err)
	}

	return nil
}

func (s *PostgresStore) listReminderRules(ctx context.Context, eventID string) ([]ReminderRule, error) {
	const query = `
		SELECT id, event_id, offset_minutes, enabled, channel, importance, created_at, updated_at
		FROM reminder_rules
		WHERE event_id = $1
		ORDER BY offset_minutes ASC, id ASC
	`

	rows, err := s.pool.Query(ctx, query, eventID)
	if err != nil {
		return nil, fmt.Errorf("list reminder rules: %w", err)
	}
	defer rows.Close()

	rules := make([]ReminderRule, 0)
	for rows.Next() {
		rule, err := scanReminderRule(rows)
		if err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate reminder rules: %w", err)
	}

	return rules, nil
}

func scanReminderRule(scanner eventScanner) (ReminderRule, error) {
	var rule ReminderRule
	if err := scanner.Scan(
		&rule.ID,
		&rule.EventID,
		&rule.OffsetMinutes,
		&rule.Enabled,
		&rule.Channel,
		&rule.Importance,
		&rule.CreatedAt,
		&rule.UpdatedAt,
	); err != nil {
		return ReminderRule{}, fmt.Errorf("scan reminder rule: %w", err)
	}

	return rule, nil
}

func scanReminderJob(scanner eventScanner) (ReminderJob, error) {
	var job ReminderJob
	if err := scanner.Scan(
		&job.ID,
		&job.UserID,
		&job.EventID,
		&job.ReminderRuleID,
		&job.EventTitle,
		&job.EventStartsAt,
		&job.OffsetMinutes,
		&job.Channel,
		&job.Importance,
		&job.Status,
		&job.ScheduledAt,
		&job.SentAt,
		&job.ReadAt,
		&job.DismissedAt,
		&job.SnoozedFromID,
		&job.SnoozedAt,
		&job.CancelledAt,
		&job.ReminderGeneration,
		&job.CreatedAt,
		&job.UpdatedAt,
	); err != nil {
		return ReminderJob{}, fmt.Errorf("scan reminder job: %w", err)
	}

	return job, nil
}

func scanEventRecipient(scanner eventScanner) (EventRecipient, error) {
	var recipient EventRecipient
	if err := scanner.Scan(
		&recipient.EventID,
		&recipient.UserID,
		&recipient.SourceType,
		&recipient.SourceID,
		&recipient.CreatedAt,
	); err != nil {
		return EventRecipient{}, fmt.Errorf("scan event recipient: %w", err)
	}

	return recipient, nil
}

func rollback(ctx context.Context, tx pgx.Tx) {
	_ = tx.Rollback(ctx)
}
