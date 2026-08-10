package device

import "time"

type Platform string

const (
	PlatformAndroid Platform = "android"
	PlatformIOS     Platform = "ios"
	PlatformWeb     Platform = "web"
)

type Device struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	Platform   Platform  `json:"platform"`
	PushToken  string    `json:"push_token"`
	Enabled    bool      `json:"enabled"`
	LastSeenAt time.Time `json:"last_seen_at"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type RegisterInput struct {
	UserID    string
	Platform  Platform
	PushToken string
}
