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

func TestCORSAllowsVolunteerUpdate(t *testing.T) {
	handler := withCORS(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Fatal("preflight request should not reach the next handler")
	}))
	request := httptest.NewRequest(http.MethodOptions, "/api/volunteers/vol_123", nil)
	request.Header.Set("Origin", "http://localhost:5174")
	request.Header.Set("Access-Control-Request-Method", http.MethodPut)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, response.Code)
	}
	if !strings.Contains(response.Header().Get("Access-Control-Allow-Methods"), http.MethodPut) {
		t.Fatalf("expected PUT in allowed methods, got %q", response.Header().Get("Access-Control-Allow-Methods"))
	}
}

func TestCORSAllowsConfiguredProductionOrigin(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://app.example.com, https://admin.example.com/")
	handler := withCORS(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for _, origin := range []string{"https://app.example.com", "https://admin.example.com"} {
		request := httptest.NewRequest(http.MethodOptions, "/api/volunteers", nil)
		request.Header.Set("Origin", origin)
		request.Header.Set("Access-Control-Request-Method", http.MethodGet)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)

		if response.Header().Get("Access-Control-Allow-Origin") != origin {
			t.Fatalf("expected configured origin %q to be allowed, got headers %#v", origin, response.Header())
		}
	}
}

func TestCORSRejectsUnconfiguredProductionOrigin(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://app.example.com")
	handler := withCORS(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	request := httptest.NewRequest(http.MethodOptions, "/api/volunteers", nil)
	request.Header.Set("Origin", "https://app.example.com.evil.test")
	request.Header.Set("Access-Control-Request-Method", http.MethodGet)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if origin := response.Header().Get("Access-Control-Allow-Origin"); origin != "" {
		t.Fatalf("expected unconfigured origin to be rejected, got %q", origin)
	}
}
