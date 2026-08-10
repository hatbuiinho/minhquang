package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"reminder/be/internal/account"
	"reminder/be/internal/device"
	"reminder/be/internal/event"
	"reminder/be/internal/ota"
)

func TestReminderJobInboxReadAndDismissRoutes(t *testing.T) {
	now := time.Date(2026, 8, 2, 10, 0, 0, 0, time.UTC)
	eventService := event.NewService(event.NewMemoryStore(), func() time.Time { return now })
	router := NewRouter(
		eventService,
		device.NewService(device.NewMemoryStore(), func() time.Time { return now }),
		account.NewService(account.NewMemoryStore(), func() time.Time { return now }),
		ota.NewService(t.TempDir()),
		t.TempDir(),
	)

	createBody := []byte(`{
		"title": "Hộ chiếu",
		"starts_at": "2026-08-02T11:00:00Z",
		"reminders": [{"offset_minutes": 60}]
	}`)
	createRequest := httptest.NewRequest(http.MethodPost, "/api/events", bytes.NewReader(createBody))
	createResponse := httptest.NewRecorder()
	router.ServeHTTP(createResponse, createRequest)
	if createResponse.Code != http.StatusCreated {
		t.Fatalf("expected create status %d, got %d: %s", http.StatusCreated, createResponse.Code, createResponse.Body.String())
	}

	listRequest := httptest.NewRequest(http.MethodGet, "/api/reminder-jobs?limit=10", nil)
	listResponse := httptest.NewRecorder()
	router.ServeHTTP(listResponse, listRequest)
	if listResponse.Code != http.StatusOK {
		t.Fatalf("expected list status %d, got %d: %s", http.StatusOK, listResponse.Code, listResponse.Body.String())
	}

	var listBody struct {
		Reminders []event.ReminderJob `json:"reminders"`
	}
	if err := json.NewDecoder(listResponse.Body).Decode(&listBody); err != nil {
		t.Fatalf("decode reminders: %v", err)
	}
	if len(listBody.Reminders) != 1 {
		t.Fatalf("expected one due reminder, got %d", len(listBody.Reminders))
	}

	readRequest := httptest.NewRequest(http.MethodPost, "/api/reminder-jobs/"+listBody.Reminders[0].ID+"/read", nil)
	readResponse := httptest.NewRecorder()
	router.ServeHTTP(readResponse, readRequest)
	if readResponse.Code != http.StatusOK {
		t.Fatalf("expected read status %d, got %d: %s", http.StatusOK, readResponse.Code, readResponse.Body.String())
	}
	var readBody event.ReminderJob
	if err := json.NewDecoder(readResponse.Body).Decode(&readBody); err != nil {
		t.Fatalf("decode read reminder: %v", err)
	}
	if readBody.ReadAt == nil {
		t.Fatal("expected read_at to be set")
	}

	dismissRequest := httptest.NewRequest(http.MethodPost, "/api/reminder-jobs/"+listBody.Reminders[0].ID+"/dismiss", nil)
	dismissResponse := httptest.NewRecorder()
	router.ServeHTTP(dismissResponse, dismissRequest)
	if dismissResponse.Code != http.StatusOK {
		t.Fatalf("expected dismiss status %d, got %d: %s", http.StatusOK, dismissResponse.Code, dismissResponse.Body.String())
	}

	listAfterDismissResponse := httptest.NewRecorder()
	router.ServeHTTP(listAfterDismissResponse, listRequest)
	if listAfterDismissResponse.Code != http.StatusOK {
		t.Fatalf("expected list after dismiss status %d, got %d", http.StatusOK, listAfterDismissResponse.Code)
	}
	listBody.Reminders = nil
	if err := json.NewDecoder(listAfterDismissResponse.Body).Decode(&listBody); err != nil {
		t.Fatalf("decode reminders after dismiss: %v", err)
	}
	if len(listBody.Reminders) != 0 {
		t.Fatalf("expected no reminders after dismiss, got %d", len(listBody.Reminders))
	}
}

func TestReminderJobSnoozeRoute(t *testing.T) {
	now := time.Date(2026, 8, 2, 10, 0, 0, 0, time.UTC)
	eventService := event.NewService(event.NewMemoryStore(), func() time.Time { return now })
	router := NewRouter(
		eventService,
		device.NewService(device.NewMemoryStore(), func() time.Time { return now }),
		account.NewService(account.NewMemoryStore(), func() time.Time { return now }),
		ota.NewService(t.TempDir()),
		t.TempDir(),
	)

	createBody := []byte(`{
		"title": "Hộ chiếu",
		"starts_at": "2026-08-02T11:00:00Z",
		"reminders": [{"offset_minutes": 60}]
	}`)
	createRequest := httptest.NewRequest(http.MethodPost, "/api/events", bytes.NewReader(createBody))
	createResponse := httptest.NewRecorder()
	router.ServeHTTP(createResponse, createRequest)
	if createResponse.Code != http.StatusCreated {
		t.Fatalf("expected create status %d, got %d: %s", http.StatusCreated, createResponse.Code, createResponse.Body.String())
	}

	listRequest := httptest.NewRequest(http.MethodGet, "/api/reminder-jobs?limit=10", nil)
	listResponse := httptest.NewRecorder()
	router.ServeHTTP(listResponse, listRequest)
	if listResponse.Code != http.StatusOK {
		t.Fatalf("expected list status %d, got %d: %s", http.StatusOK, listResponse.Code, listResponse.Body.String())
	}
	var listBody struct {
		Reminders []event.ReminderJob `json:"reminders"`
	}
	if err := json.NewDecoder(listResponse.Body).Decode(&listBody); err != nil {
		t.Fatalf("decode reminders: %v", err)
	}
	if len(listBody.Reminders) != 1 {
		t.Fatalf("expected one due reminder, got %d", len(listBody.Reminders))
	}

	snoozeBody := []byte(`{"delay_minutes": 15}`)
	snoozeRequest := httptest.NewRequest(
		http.MethodPost,
		"/api/reminder-jobs/"+listBody.Reminders[0].ID+"/snooze",
		bytes.NewReader(snoozeBody),
	)
	snoozeResponse := httptest.NewRecorder()
	router.ServeHTTP(snoozeResponse, snoozeRequest)
	if snoozeResponse.Code != http.StatusOK {
		t.Fatalf("expected snooze status %d, got %d: %s", http.StatusOK, snoozeResponse.Code, snoozeResponse.Body.String())
	}
	var snoozed event.ReminderJob
	if err := json.NewDecoder(snoozeResponse.Body).Decode(&snoozed); err != nil {
		t.Fatalf("decode snoozed reminder: %v", err)
	}
	if snoozed.SnoozedFromID != listBody.Reminders[0].ID {
		t.Fatalf("expected snoozed_from_id, got %+v", snoozed)
	}
}
