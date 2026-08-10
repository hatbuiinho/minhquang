package httpapi

import "testing"

func TestAllowedOriginAllowsLocalAndPrivateDevelopmentNetworks(t *testing.T) {
	allowed := []string{
		"http://localhost:5173",
		"http://127.0.0.1:5173",
		"http://10.10.10.198:5173",
		"http://172.20.1.10:5173",
		"http://192.168.1.20:5173",
		"http://100.98.122.91:5173",
	}

	for _, origin := range allowed {
		if !allowedOrigin(origin) {
			t.Fatalf("expected origin %q to be allowed", origin)
		}
	}
}

func TestAllowedOriginRejectsPublicHosts(t *testing.T) {
	blocked := []string{
		"https://example.com",
		"http://8.8.8.8:5173",
		"not a url",
	}

	for _, origin := range blocked {
		if allowedOrigin(origin) {
			t.Fatalf("expected origin %q to be rejected", origin)
		}
	}
}
