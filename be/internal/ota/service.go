package ota

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const defaultChannel = "dev"

var ErrInvalidInput = errors.New("invalid ota input")

type Service struct {
	storageDir string
}

func NewService(storageDir string) *Service {
	storageDir = strings.TrimSpace(storageDir)
	if storageDir == "" {
		storageDir = filepath.Join("storage", "ota")
	}

	return &Service{storageDir: storageDir}
}

func (s *Service) Latest(_ context.Context, input CheckInput) (Update, error) {
	platform := sanitizeSegment(input.Platform)
	if platform == "" {
		return Update{}, fmt.Errorf("%w: platform is required", ErrInvalidInput)
	}

	channel := sanitizeSegment(input.Channel)
	if channel == "" {
		channel = defaultChannel
	}

	file := filepath.Join(s.storageDir, platform, channel, "latest.json")
	content, err := os.ReadFile(file)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Update{Available: false, Platform: platform, Channel: channel}, nil
		}
		return Update{}, fmt.Errorf("read ota metadata: %w", err)
	}

	var update Update
	if err := json.Unmarshal(content, &update); err != nil {
		return Update{}, fmt.Errorf("parse ota metadata: %w", err)
	}

	update.Platform = platform
	update.Channel = channel
	update.Version = strings.TrimSpace(update.Version)
	update.URL = strings.TrimSpace(update.URL)
	update.Checksum = strings.TrimSpace(update.Checksum)
	update.MinNativeVersion = strings.TrimSpace(update.MinNativeVersion)
	update.MaxNativeVersion = strings.TrimSpace(update.MaxNativeVersion)

	if update.Version == "" || update.URL == "" {
		return Update{}, fmt.Errorf("%w: ota metadata requires version and url", ErrInvalidInput)
	}
	if !nativeVersionAllowed(input.NativeVersion, update) ||
		strings.TrimSpace(input.CurrentVersion) == update.Version {
		return Update{Available: false, Platform: platform, Channel: channel}, nil
	}

	update.Available = true
	return update, nil
}

func sanitizeSegment(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	for _, item := range value {
		if item >= 'a' && item <= 'z' {
			continue
		}
		if item >= '0' && item <= '9' {
			continue
		}
		if item == '-' || item == '_' {
			continue
		}
		return ""
	}
	return value
}

func nativeVersionAllowed(nativeVersion string, update Update) bool {
	nativeVersion = strings.TrimSpace(nativeVersion)
	if update.MinNativeVersion != "" && nativeVersion < update.MinNativeVersion {
		return false
	}
	if update.MaxNativeVersion != "" && nativeVersion > update.MaxNativeVersion {
		return false
	}
	return true
}
