package structs

type Customer struct {
	ID      int `json:"id"`
	Balance int `json:"balance"`
	Limit   int `json:"limit"`
}
