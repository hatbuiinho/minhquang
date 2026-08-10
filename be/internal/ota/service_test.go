package ota

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestServiceLatestReturnsUnavailableWhenNoMetadataExists(t *testing.T) {
	service := NewService(t.TempDir())

	update, err := service.Latest(context.Background(), CheckInput{
		Platform: "android",
	})
	if err != nil {
		t.Fatalf("latest update: %v", err)
	}
	if update.Available {
		t.Fatal("expected update to be unavailable")
	}
	if update.Channel != defaultChannel {
		t.Fatalf("expected default channel, got %q", update.Channel)
	}
}

func TestServiceLatestReturnsUpdateWhenVersionDiffers(t *testing.T) {
	dir := t.TempDir()
	writeMetadata(t, dir, "android", "dev", `{
		"version": "2026.08.03.001",
		"url": "/ota/android/dev/2026.08.03.001.zip",
		"checksum": "sha256-example",
		"mandatory": true,
		"min_native_version": "1.0",
		"max_native_version": "1.0"
	}`)
	service := NewService(dir)

	update, err := service.Latest(context.Background(), CheckInput{
		Platform:       "android",
		Channel:        "dev",
		CurrentVersion: "builtin",
		NativeVersion:  "1.0",
	})
	if err != nil {
		t.Fatalf("latest update: %v", err)
	}
	if !update.Available {
		t.Fatal("expected update to be available")
	}
	if !update.Mandatory {
		t.Fatal("expected update to be mandatory")
	}
	if update.Version != "2026.08.03.001" {
		t.Fatalf("unexpected version %q", update.Version)
	}
}

func TestServiceLatestReturnsUnavailableForSameVersion(t *testing.T) {
	dir := t.TempDir()
	writeMetadata(t, dir, "android", "dev", `{
		"version": "2026.08.03.001",
		"url": "/ota/android/dev/2026.08.03.001.zip"
	}`)
	service := NewService(dir)

	update, err := service.Latest(context.Background(), CheckInput{
		Platform:       "android",
		CurrentVersion: "2026.08.03.001",
	})
	if err != nil {
		t.Fatalf("latest update: %v", err)
	}
	if update.Available {
		t.Fatal("expected same version to be unavailable")
	}
}

func TestServiceLatestRejectsUnsafeSegments(t *testing.T) {
	service := NewService(t.TempDir())

	_, err := service.Latest(context.Background(), CheckInput{
		Platform: "../android",
	})
	if err == nil {
		t.Fatal("expected invalid platform error")
	}
}

func writeMetadata(t *testing.T, dir string, platform string, channel string, content string) {
	t.Helper()
	targetDir := filepath.Join(dir, platform, channel)
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		t.Fatalf("mkdir metadata dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(targetDir, "latest.json"), []byte(content), 0o644); err != nil {
		t.Fatalf("write metadata: %v", err)
	}
}
