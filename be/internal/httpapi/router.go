package httpapi

import (
	"net/http"
	"net/netip"
	"net/url"

	"minhquang/be/internal/department"
	"minhquang/be/internal/device"
	"minhquang/be/internal/docs"
	"minhquang/be/internal/ota"
	"minhquang/be/internal/user"
	"minhquang/be/internal/volunteer"
)

func NewRouter(
	devices *device.Service,
	users *user.Service,
	volunteers *volunteer.Service,
	departments *department.Service,
	updates *ota.Service,
	otaStorageDir string,
) http.Handler {
	mux := http.NewServeMux()
	deviceHandler := NewDeviceHandler(devices)
	authHandler := NewAuthHandler(users)
	userHandler := NewUserHandler(users)
	volunteerHandler := NewVolunteerHandler(volunteers)
	departmentHandler := NewDepartmentHandler(departments)
	otaHandler := NewOTAHandler(updates)

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /openapi.yaml", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(docs.OpenAPIYAML)
	})
	mux.HandleFunc("POST /api/auth/login", authHandler.Login)
	protected := http.NewServeMux()
	protected.HandleFunc("/api/auth/logout", authHandler.Logout)
	protected.HandleFunc("/api/auth/me", authHandler.Me)
	protected.HandleFunc("/api/users", userHandler.Collection)
	protected.HandleFunc("/api/volunteers", volunteerHandler.Collection)
	protected.HandleFunc("/api/volunteers/", volunteerHandler.Item)
	protected.HandleFunc("/api/departments", departmentHandler.Collection)
	protected.HandleFunc("/api/departments/", departmentHandler.Item)
	protected.HandleFunc("GET /api/volunteer-options/departments", departmentHandler.Options)
	protected.HandleFunc("/api/devices", deviceHandler.Collection)
	mux.Handle("/api/", requireAuth(users, protected))
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
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
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
