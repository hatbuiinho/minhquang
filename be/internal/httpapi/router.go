package httpapi

import (
	"net/http"
	"net/netip"
	"net/url"

	"reminder/be/internal/account"
	"reminder/be/internal/device"
	"reminder/be/internal/docs"
	"reminder/be/internal/event"
	"reminder/be/internal/ota"
)

func NewRouter(
	events *event.Service,
	devices *device.Service,
	accounts *account.Service,
	updates *ota.Service,
	otaStorageDir string,
) http.Handler {
	mux := http.NewServeMux()
	handler := NewEventHandler(events)
	deviceHandler := NewDeviceHandler(devices)
	accountHandler := NewAccountHandler(accounts)
	otaHandler := NewOTAHandler(updates)

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /openapi.yaml", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(docs.OpenAPIYAML)
	})
	mux.HandleFunc("/api/events", handler.Collection)
	mux.HandleFunc("/api/events/", handler.Item)
	mux.HandleFunc("/api/devices", deviceHandler.Collection)
	mux.HandleFunc("/api/users", accountHandler.Users)
	mux.HandleFunc("/api/groups", accountHandler.Groups)
	mux.HandleFunc("/api/groups/", accountHandler.GroupMembers)
	mux.HandleFunc("/api/reminders/upcoming", handler.UpcomingReminders)
	mux.HandleFunc("/api/reminder-jobs", handler.ReminderJobs)
	mux.HandleFunc("/api/reminder-jobs/", handler.ReminderJobItem)
	mux.HandleFunc("/api/app-updates/android/latest", otaHandler.AndroidLatest)
	mux.Handle("/ota/", http.StripPrefix("/ota/", http.FileServer(http.Dir(otaStorageDir))))

	return withRequestLog(withCORS(mux))
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && allowedOrigin(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-User-ID")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func allowedOrigin(origin string) bool {
	parsed, err := url.Parse(origin)
	if err != nil {
		return false
	}

	host := parsed.Hostname()
	if host == "localhost" {
		return true
	}

	addr, err := netip.ParseAddr(host)
	if err != nil {
		return false
	}

	for _, prefix := range allowedDevOriginPrefixes() {
		if prefix.Contains(addr) {
			return true
		}
	}

	return false
}

func allowedDevOriginPrefixes() []netip.Prefix {
	return []netip.Prefix{
		netip.MustParsePrefix("127.0.0.0/8"),
		netip.MustParsePrefix("10.0.0.0/8"),
		netip.MustParsePrefix("172.16.0.0/12"),
		netip.MustParsePrefix("192.168.0.0/16"),
		netip.MustParsePrefix("100.64.0.0/10"),
	}
}
