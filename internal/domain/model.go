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

/**
  {
	  "order": "2377225624",
	  "sum": 500,
	  "processed_at": "2020-12-09T16:09:57+03:00"
  }
*/

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
