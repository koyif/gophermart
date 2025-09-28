package dto

/**
{
    "order": "2377225624",
    "sum": 751
}
*/

type Withdrawal struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}
