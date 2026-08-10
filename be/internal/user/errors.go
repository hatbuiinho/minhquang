package user

import "errors"

var (
	ErrInvalidInput       = errors.New("invalid user input")
	ErrNotFound           = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUsernameExists     = errors.New("username already exists")
)
