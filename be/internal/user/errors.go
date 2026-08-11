package user

import "errors"

var (
	ErrInvalidInput       = errors.New("invalid user input")
	ErrNotFound           = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrCurrentPassword    = errors.New("current password is incorrect")
	ErrUsernameExists     = errors.New("username already exists")
)
