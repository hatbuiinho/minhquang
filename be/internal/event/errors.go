package event

import "errors"

var (
	ErrNotFound      = errors.New("event not found")
	ErrInvalidInput  = errors.New("invalid event input")
	ErrInvalidStatus = errors.New("invalid event status")
	ErrInvalidRule   = errors.New("invalid reminder rule")
)
