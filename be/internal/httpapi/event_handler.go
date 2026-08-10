package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"reminder/be/internal/event"
)

const defaultUserID = "local-user"

type EventHandler struct {
	events *event.Service
}

type eventPayload struct {
	Title             *string                `json:"title"`
	Description       *string                `json:"description"`
	StartsAt          *apiTime               `json:"starts_at"`
	Timezone          *string                `json:"timezone"`
	Status            *event.Status          `json:"status"`
	Reminders         *[]reminderRulePayload `json:"reminders"`
	AudienceType      *event.AudienceType    `json:"audience_type"`
	RecipientUserIDs  *[]string              `json:"recipient_user_ids"`
	RecipientGroupIDs *[]string              `json:"recipient_group_ids"`
}

type reminderRulePayload struct {
	OffsetMinutes int                       `json:"offset_minutes"`
	Enabled       *bool                     `json:"enabled"`
	Channel       *string                   `json:"channel"`
	Importance    *event.ReminderImportance `json:"importance"`
}

type reminderSnoozePayload struct {
	DelayMinutes int `json:"delay_minutes"`
}

type apiTime struct {
	time.Time
}

func NewEventHandler(events *event.Service) *EventHandler {
	return &EventHandler{events: events}
}

func (h *EventHandler) Collection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.list(w, r)
	case http.MethodPost:
		h.create(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
	}
}

func (h *EventHandler) Item(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/events/")
	if id == "" || strings.Contains(id, "/") {
		writeError(w, http.StatusNotFound, "not_found", "event not found")
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.get(w, r, id)
	case http.MethodPatch:
		h.update(w, r, id)
	case http.MethodDelete:
		h.delete(w, r, id)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
	}
}

func (h *EventHandler) UpcomingReminders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
		return
	}

	limit := 50
	if value := strings.TrimSpace(r.URL.Query().Get("limit")); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 1 || parsed > 100 {
			writeError(w, http.StatusBadRequest, "invalid_input", "limit must be between 1 and 100")
			return
		}
		limit = parsed
	}

	items, err := h.events.ListUpcomingReminderJobs(r.Context(), userID(r), limit)
	if err != nil {
		handleEventError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string][]event.ReminderJob{"reminders": items})
}

func (h *EventHandler) ReminderJobs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
		return
	}

	limit, ok := reminderJobLimit(w, r)
	if !ok {
		return
	}

	items, err := h.events.ListReminderInbox(r.Context(), userID(r), limit)
	if err != nil {
		handleEventError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string][]event.ReminderJob{"reminders": items})
}

func (h *EventHandler) ReminderJobItem(w http.ResponseWriter, r *http.Request) {
	trimmed := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/reminder-jobs/"), "/")
	parts := strings.Split(trimmed, "/")
	if len(parts) != 2 || parts[0] == "" {
		writeError(w, http.StatusNotFound, "not_found", "reminder job not found")
		return
	}

	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
		return
	}

	var (
		item event.ReminderJob
		err  error
	)
	switch parts[1] {
	case "read":
		item, err = h.events.MarkReminderJobRead(r.Context(), userID(r), parts[0])
	case "dismiss":
		item, err = h.events.DismissReminderJob(r.Context(), userID(r), parts[0])
	case "snooze":
		var payload reminderSnoozePayload
		if err := readJSON(r, &payload); err != nil {
			writeError(w, http.StatusBadRequest, "invalid_json", "request body must be valid json")
			return
		}
		item, err = h.events.SnoozeReminderJob(r.Context(), userID(r), parts[0], payload.DelayMinutes)
	default:
		writeError(w, http.StatusNotFound, "not_found", "reminder action not found")
		return
	}
	if err != nil {
		handleEventError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, item)
}

func (h *EventHandler) list(w http.ResponseWriter, r *http.Request) {
	limit := 20
	if value := strings.TrimSpace(r.URL.Query().Get("limit")); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 1 || parsed > 100 {
			writeError(w, http.StatusBadRequest, "invalid_input", "limit must be between 1 and 100")
			return
		}
		limit = parsed
	}

	page, err := h.events.List(r.Context(), userID(r), event.ListOptions{
		Limit:  limit,
		Cursor: r.URL.Query().Get("cursor"),
	})
	if err != nil {
		handleEventError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, page)
}

func reminderJobLimit(w http.ResponseWriter, r *http.Request) (int, bool) {
	limit := 50
	if value := strings.TrimSpace(r.URL.Query().Get("limit")); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 1 || parsed > 100 {
			writeError(w, http.StatusBadRequest, "invalid_input", "limit must be between 1 and 100")
			return 0, false
		}
		limit = parsed
	}

	return limit, true
}

func (h *EventHandler) create(w http.ResponseWriter, r *http.Request) {
	var payload eventPayload
	if err := readJSON(r, &payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "request body must be valid json")
		return
	}
	if payload.Title == nil || payload.StartsAt == nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "title and starts_at are required")
		return
	}

	input := event.CreateInput{
		UserID:   userID(r),
		Title:    *payload.Title,
		StartsAt: payload.StartsAt.Time,
	}
	if payload.Description != nil {
		input.Description = *payload.Description
	}
	if payload.Timezone != nil {
		input.Timezone = *payload.Timezone
	}
	if payload.Reminders != nil {
		input.Reminders = reminderRuleInputs(*payload.Reminders)
	}
	if payload.AudienceType != nil {
		input.AudienceType = *payload.AudienceType
	}
	if payload.RecipientUserIDs != nil {
		input.RecipientUserIDs = *payload.RecipientUserIDs
	}
	if payload.RecipientGroupIDs != nil {
		input.RecipientGroupIDs = *payload.RecipientGroupIDs
	}

	item, err := h.events.Create(r.Context(), input)
	if err != nil {
		handleEventError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, item)
}

func (h *EventHandler) get(w http.ResponseWriter, r *http.Request, id string) {
	item, err := h.events.Get(r.Context(), userID(r), id)
	if err != nil {
		handleEventError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, item)
}

func (h *EventHandler) update(w http.ResponseWriter, r *http.Request, id string) {
	var payload eventPayload
	if err := readJSON(r, &payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "request body must be valid json")
		return
	}

	input := event.UpdateInput{
		Title:             payload.Title,
		Description:       payload.Description,
		Timezone:          payload.Timezone,
		Status:            payload.Status,
		AudienceType:      payload.AudienceType,
		RecipientUserIDs:  payload.RecipientUserIDs,
		RecipientGroupIDs: payload.RecipientGroupIDs,
	}
	if payload.StartsAt != nil {
		input.StartsAt = &payload.StartsAt.Time
	}
	if payload.Reminders != nil {
		reminders := reminderRuleInputs(*payload.Reminders)
		input.Reminders = &reminders
	}

	item, err := h.events.Update(r.Context(), userID(r), id, input)
	if err != nil {
		handleEventError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, item)
}

func (h *EventHandler) delete(w http.ResponseWriter, r *http.Request, id string) {
	if err := h.events.Delete(r.Context(), userID(r), id); err != nil {
		handleEventError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (t *apiTime) UnmarshalJSON(data []byte) error {
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return err
	}

	t.Time = parsed
	return nil
}

func userID(r *http.Request) string {
	value := strings.TrimSpace(r.Header.Get("X-User-ID"))
	if value == "" {
		return defaultUserID
	}
	return value
}

func handleEventError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, event.ErrNotFound):
		writeError(w, http.StatusNotFound, "not_found", "event not found")
	case errors.Is(err, event.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, "invalid_input", err.Error())
	case errors.Is(err, event.ErrInvalidStatus):
		writeError(w, http.StatusBadRequest, "invalid_status", "status is invalid")
	case errors.Is(err, event.ErrInvalidRule):
		writeError(w, http.StatusBadRequest, "invalid_reminder_rule", err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}

func reminderRuleInputs(payloads []reminderRulePayload) []event.ReminderRuleInput {
	inputs := make([]event.ReminderRuleInput, len(payloads))
	for index, payload := range payloads {
		inputs[index] = event.ReminderRuleInput{
			OffsetMinutes: payload.OffsetMinutes,
			Enabled:       payload.Enabled,
		}
		if payload.Channel != nil {
			inputs[index].Channel = *payload.Channel
		}
		if payload.Importance != nil {
			inputs[index].Importance = *payload.Importance
		}
	}

	return inputs
}
