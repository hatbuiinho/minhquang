package department

import "errors"

var (
	ErrInvalidInput = errors.New("invalid department input")
	ErrNotFound     = errors.New("department not found")
	ErrNameExists   = errors.New("department name already exists")
	ErrInactive     = errors.New("department is inactive")
	ErrInUse        = errors.New("department is in use")
)
