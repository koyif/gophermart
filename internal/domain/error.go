package domain

import "errors"

var (
	ErrUserExists              = errors.New("user already exists")
	ErrIncorrectCredentials    = errors.New("incorrect credentials")
	ErrOrderExists             = errors.New("order already exists")
	ErrOrderAddedByAnotherUser = errors.New("order added by another user")
)
