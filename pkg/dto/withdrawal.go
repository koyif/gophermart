package dto

/**
{
  "order": "2377225624",
  "sum": 500,
  "processed_at": "2020-12-09T16:09:57+03:00"
}
*/

type Withdrawal struct {
	Order       string  `json:"order"`
	Sum         float64 `json:"sum"`
	ProcessedAt string  `json:"processed_at,omitempty"`
}
