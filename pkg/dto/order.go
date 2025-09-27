package dto

/**
  {
      "number": "9278923470",
      "status": "PROCESSED",
      "accrual": 500,
      "uploaded_at": "2020-12-10T15:15:45+03:00"
  },
*/

type Order struct {
	Number     string `json:"number"`
	Status     string `json:"status"`
	Accrual    *int64 `json:"accrual,omitempty"`
	UploadedAt string `json:"uploaded_at"`
}
