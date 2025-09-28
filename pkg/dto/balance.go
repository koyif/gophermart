package dto

/**
  {
      "current": 500.5,
      "withdrawn": 42
  }
*/

type Balance struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}
