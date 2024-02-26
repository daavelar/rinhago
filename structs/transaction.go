package structs

import "time"

type Transaction struct {
	ID          int       `json:"id"`
	CustomerID  int       `json:"customer_id"`
	Value       int       `json:"value"`
	Type        string    `json:"type"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}
