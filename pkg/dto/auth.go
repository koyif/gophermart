package dto

import (
	"errors"
	"fmt"
	"strings"
)

type Auth struct {
	Login    string
	Password string
}

func (a Auth) IsValid() error {
	var logingErr, passwordErr error

	if strings.TrimSpace(a.Login) == "" {
		logingErr = fmt.Errorf("login is required")
	}

	if strings.TrimSpace(a.Password) == "" {
		passwordErr = fmt.Errorf("password is required")
	}

	return errors.Join(logingErr, passwordErr)
}
