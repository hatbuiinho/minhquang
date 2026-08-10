package account

import "errors"

var (
	ErrInvalidInput = errors.New("invalid account input")
	ErrNotFound     = errors.New("account resource not found")
)
