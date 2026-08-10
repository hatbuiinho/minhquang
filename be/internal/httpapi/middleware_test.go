package httpapi

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRequestLogAddsRequestIDAndLogsResponse(t *testing.T) {
	var logs bytes.Buffer
	previousOutput := log.Writer()
	log.SetOutput(&logs)
	defer log.SetOutput(previousOutput)

	handler := withRequestLog(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(requestIDHeader) != "client-request-id" {
			t.Fatalf("expected incoming request id to be preserved")
		}
		writeJSON(w, http.StatusCreated, map[string]string{"status": "created"})
	}))

	request := httptest.NewRequest(http.MethodPost, "/api/events", nil)
	request.Header.Set(requestIDHeader, "client-request-id")
	request.Header.Set("User-Agent", "reminder-test")
	request.Header.Set("X-Forwarded-For", "203.0.113.10, 10.0.0.2")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, response.Code)
	}
	if response.Header().Get(requestIDHeader) != "client-request-id" {
		t.Fatalf("expected response request id header")
	}

	output := logs.String()
	for _, expected := range []string{
		"http_request",
		"method=POST",
		"path=/api/events",
		"status=201",
		"remote_ip=203.0.113.10",
		"request_id=client-request-id",
		"user_agent=\"reminder-test\"",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("expected log to contain %q, got %q", expected, output)
		}
	}
}

func TestRequestLogSkipsHealthAndOptions(t *testing.T) {
	var logs bytes.Buffer
	previousOutput := log.Writer()
	log.SetOutput(&logs)
	defer log.SetOutput(previousOutput)

	handler := withRequestLog(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	for _, request := range []*http.Request{
		httptest.NewRequest(http.MethodGet, "/healthz", nil),
		httptest.NewRequest(http.MethodOptions, "/api/events", nil),
	} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Header().Get(requestIDHeader) == "" {
			t.Fatalf("expected request id header")
		}
	}

	if logs.Len() != 0 {
		t.Fatalf("expected no request log, got %q", logs.String())
	}
}
