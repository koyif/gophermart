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
	Accrual    *float64
	UploadedAt time.Time
}

type Withdrawal struct {
	UserID      int64
	OrderNumber string
	Amount      float64
	ProcessedAt time.Time
}

type Balance struct {
	Current   float64
	Withdrawn float64
}
