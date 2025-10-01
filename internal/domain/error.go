package domain

import "errors"

var (
	ErrUserNotFound                 = errors.New("user not found")
	ErrUserExists                   = errors.New("user already exists")
	ErrIncorrectCredentials         = errors.New("incorrect credentials")
	ErrOrderExists                  = errors.New("order already exists")
	ErrOrderAddedByAnotherUser      = errors.New("order added by another user")
	ErrWithdrawalExists             = errors.New("withdrawal already exists")
	ErrWithdrawalAddedByAnotherUser = errors.New("withdrawal added by another user")
	ErrInsufficientFunds            = errors.New("insufficient funds")
)
