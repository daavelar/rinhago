package structs

type TransactionPayload struct {
	Value       int    `json:"valor" binding:"required"`
	Type        string `json:"tipo" binding:"required,oneof=c d"`
	Description string `json:"descricao"`
}
