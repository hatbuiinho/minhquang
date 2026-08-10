package push

import "context"

type Importance string

const (
	ImportanceNormal Importance = "normal"
	ImportanceUrgent Importance = "urgent"
)

type Message struct {
	Token      string
	Title      string
	Body       string
	Importance Importance
	Data       map[string]string
}

type Sender interface {
	Send(ctx context.Context, message Message) error
}
