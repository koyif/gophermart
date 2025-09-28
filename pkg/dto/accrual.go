package dto

/**
  {
      "order": "<number>",
      "status": "PROCESSED",
      "accrual": 500
  }
*/

type AccrualResponse struct {
	Order   string   `json:"order"`
	Status  string   `json:"status"`
	Accrual *float64 `json:"accrual,omitempty"`
}
