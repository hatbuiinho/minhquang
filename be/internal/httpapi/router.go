package httpapi

import (
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"strings"

	"minhquang/be/internal/department"
	"minhquang/be/internal/device"
	"minhquang/be/internal/docs"
	"minhquang/be/internal/ota"
	"minhquang/be/internal/storage"
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
	objectStorage *storage.MinIO,
) http.Handler {
	mux := http.NewServeMux()
	deviceHandler := NewDeviceHandler(devices)
	authHandler := NewAuthHandler(users)
	userHandler := NewUserHandler(users)
	volunteerHandler := NewVolunteerHandler(volunteers)
	departmentHandler := NewDepartmentHandler(departments)
	otaHandler := NewOTAHandler(updates)
	uploadHandler := NewUploadHandler(objectStorage)

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
	protected.HandleFunc("/api/auth/password", authHandler.ChangePassword)
	protected.HandleFunc("/api/auth/profile", authHandler.UpdateProfile)
	protected.HandleFunc("/api/auth/avatar", authHandler.UpdateAvatar)
	protected.HandleFunc("/api/uploads/presign", uploadHandler.Presign)
	protected.HandleFunc("/api/users", userHandler.Collection)
	protected.HandleFunc("/api/volunteers", volunteerHandler.Collection)
	protected.HandleFunc("/api/volunteers/bulk", volunteerHandler.Bulk)
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
	configuredOrigins := parseAllowedOrigins(os.Getenv("CORS_ALLOWED_ORIGINS"))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && allowedOrigin(origin, configuredOrigins) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Add("Vary", "Origin")
			w.Header().Add("Vary", "Access-Control-Request-Method")
			w.Header().Add("Vary", "Access-Control-Request-Headers")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Max-Age", "600")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func allowedOrigin(origin string, configuredOrigins map[string]struct{}) bool {
	normalized, parsed, ok := normalizeOrigin(origin)
	if !ok {
		return false
	}
	if _, allowed := configuredOrigins[normalized]; allowed {
		return true
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

func parseAllowedOrigins(value string) map[string]struct{} {
	origins := make(map[string]struct{})
	for _, rawOrigin := range strings.Split(value, ",") {
		origin, _, ok := normalizeOrigin(rawOrigin)
		if ok {
			origins[origin] = struct{}{}
		}
	}
	return origins
}

func normalizeOrigin(value string) (string, *url.URL, bool) {
	value = strings.TrimSpace(value)
	parsed, err := url.Parse(value)
	if err != nil || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", nil, false
	}
	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "http" && scheme != "https" {
		return "", nil, false
	}
	if parsed.Path != "" && parsed.Path != "/" {
		return "", nil, false
	}
	host := strings.ToLower(parsed.Host)
	if (scheme == "http" && parsed.Port() == "80") || (scheme == "https" && parsed.Port() == "443") {
		host = strings.ToLower(parsed.Hostname())
		if strings.Contains(host, ":") {
			host = "[" + host + "]"
		}
	}
	parsed.Scheme = scheme
	parsed.Host = host
	parsed.Path = ""
	return scheme + "://" + host, parsed, true
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
