package domain

import "time"

type User struct {
	ID           int64
	Login        string
	Password     string
	RegisteredAt time.Time
}

type Order struct {
	ID         int64
	Number     string
	UserID     int64
	Status     string
	Accrual    *int64
	UploadedAt time.Time
}
