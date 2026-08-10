package volunteer

import "errors"

var (
	ErrInvalidInput = errors.New("invalid volunteer input")
	ErrNotFound     = errors.New("volunteer not found")
)
