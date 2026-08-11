package httpapi

import "testing"

func TestAllowedAvatarType(t *testing.T) {
	for _, contentType := range []string{"image/jpeg", "image/png", "image/webp", " IMAGE/JPEG "} {
		if !allowedAvatarType(contentType) {
			t.Errorf("expected %q to be allowed", contentType)
		}
	}
	for _, contentType := range []string{"", "image/svg+xml", "text/html", "application/octet-stream"} {
		if allowedAvatarType(contentType) {
			t.Errorf("expected %q to be rejected", contentType)
		}
	}
}
