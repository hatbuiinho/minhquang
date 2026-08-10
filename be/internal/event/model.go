package event

import (
	"context"
	"time"
)

type Status string
type ReminderJobStatus string
type ReminderImportance string
type AudienceType string
type RecipientSourceType string

const (
	StatusActive    Status = "active"
	StatusArchived  Status = "archived"
	StatusCancelled Status = "cancelled"
)

const (
	AudienceSelf           AudienceType = "self"
	AudienceSelectedUsers  AudienceType = "selected_users"
	AudienceSelectedGroups AudienceType = "selected_groups"
	AudienceAllUsers       AudienceType = "all_users"
)

const (
	RecipientSourceSelf     RecipientSourceType = "self"
	RecipientSourceUser     RecipientSourceType = "user"
	RecipientSourceGroup    RecipientSourceType = "group"
	RecipientSourceAllUsers RecipientSourceType = "all_users"
)

const (
	ReminderJobPending    ReminderJobStatus = "pending"
	ReminderJobProcessing ReminderJobStatus = "processing"
	ReminderJobSent       ReminderJobStatus = "sent"
	ReminderJobCancelled  ReminderJobStatus = "cancelled"
	ReminderJobFailed     ReminderJobStatus = "failed"
)

const (
	ReminderImportanceNormal ReminderImportance = "normal"
	ReminderImportanceUrgent ReminderImportance = "urgent"
)

type Event struct {
	ID                 string           `json:"id"`
	UserID             string           `json:"user_id"`
	Title              string           `json:"title"`
	Description        string           `json:"description"`
	StartsAt           time.Time        `json:"starts_at"`
	Timezone           string           `json:"timezone"`
	Status             Status           `json:"status"`
	AudienceType       AudienceType     `json:"audience_type"`
	ReminderGeneration int              `json:"reminder_generation"`
	Reminders          []ReminderRule   `json:"reminders"`
	Recipients         []EventRecipient `json:"recipients,omitempty"`
	CreatedAt          time.Time        `json:"created_at"`
	UpdatedAt          time.Time        `json:"updated_at"`
}

type EventPage struct {
	Items      []Event `json:"events"`
	NextCursor string  `json:"next_cursor,omitempty"`
	HasMore    bool    `json:"has_more"`
}

type ListOptions struct {
	Limit  int
	Cursor string
}

type EventCursor struct {
	CreatedAt time.Time
	ID        string
}

type ReminderRule struct {
	ID            string             `json:"id"`
	EventID       string             `json:"event_id"`
	OffsetMinutes int                `json:"offset_minutes"`
	Enabled       bool               `json:"enabled"`
	Channel       string             `json:"channel"`
	Importance    ReminderImportance `json:"importance"`
	CreatedAt     time.Time          `json:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at"`
}

type EventRecipient struct {
	EventID    string              `json:"event_id"`
	UserID     string              `json:"user_id"`
	SourceType RecipientSourceType `json:"source_type"`
	SourceID   string              `json:"source_id,omitempty"`
	CreatedAt  time.Time           `json:"created_at"`
}

type ResolvedRecipient struct {
	UserID     string
	SourceType RecipientSourceType
	SourceID   string
}

type RecipientInput struct {
	AudienceType AudienceType
	UserIDs      []string
	GroupIDs     []string
}

type RecipientResolver interface {
	ResolveEventRecipients(ctx context.Context, ownerUserID string, input RecipientInput) ([]ResolvedRecipient, error)
}

type ReminderJob struct {
	ID                 string             `json:"id"`
	UserID             string             `json:"user_id"`
	EventID            string             `json:"event_id"`
	ReminderRuleID     string             `json:"reminder_rule_id"`
	EventTitle         string             `json:"event_title"`
	EventStartsAt      time.Time          `json:"event_starts_at"`
	OffsetMinutes      int                `json:"offset_minutes"`
	Channel            string             `json:"channel"`
	Importance         ReminderImportance `json:"importance"`
	Status             ReminderJobStatus  `json:"status"`
	ScheduledAt        time.Time          `json:"scheduled_at"`
	SentAt             *time.Time         `json:"sent_at,omitempty"`
	ReadAt             *time.Time         `json:"read_at,omitempty"`
	DismissedAt        *time.Time         `json:"dismissed_at,omitempty"`
	SnoozedFromID      string             `json:"snoozed_from_id,omitempty"`
	SnoozedAt          *time.Time         `json:"snoozed_at,omitempty"`
	CancelledAt        *time.Time         `json:"cancelled_at,omitempty"`
	ReminderGeneration int                `json:"reminder_generation"`
	CreatedAt          time.Time          `json:"created_at"`
	UpdatedAt          time.Time          `json:"updated_at"`
}

type CreateInput struct {
	UserID            string
	Title             string
	Description       string
	StartsAt          time.Time
	Timezone          string
	Reminders         []ReminderRuleInput
	AudienceType      AudienceType
	RecipientUserIDs  []string
	RecipientGroupIDs []string
}

type UpdateInput struct {
	Title             *string
	Description       *string
	StartsAt          *time.Time
	Timezone          *string
	Status            *Status
	Reminders         *[]ReminderRuleInput
	AudienceType      *AudienceType
	RecipientUserIDs  *[]string
	RecipientGroupIDs *[]string
}

type ReminderRuleInput struct {
	OffsetMinutes int
	Enabled       *bool
	Channel       string
	Importance    ReminderImportance
}
