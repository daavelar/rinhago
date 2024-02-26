package handlers

import (
	"net/http"
	"rinhago/structs"
	"time"

	"github.com/gin-gonic/gin"
)

func listBankStatement(c *gin.Context) {
	id := c.Param("id")

	db := database.connect()

	query := `
		SELECT transactions.id, transactions.value, transactions.type, transactions.description, transactions.created_at
		FROM transactions
		INNER JOIN customers ON transactions.customer_id = customers.id
		WHERE customers.id = ?
	`

	rows, err := db.Query(query, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao consultar o banco de dados"})
		return
	}
	defer rows.Close()

	var transactions []structs.Transaction = []structs.Transaction{}

	for rows.Next() {
		var transaction structs.Transaction
		if err := rows.Scan(&transaction.ID, &transaction.Value, &transaction.Type, &transaction.Description, &transaction.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		transactions = append(transactions, transaction)
	}

	customerQuery := "SELECT balance, `limit` FROM customers WHERE id = ?"
	var customer structs.Customer
	err = db.QueryRow(customerQuery, id).Scan(&customer.Balance, &customer.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao obter informações do cliente"})
		return
	}

	response := gin.H{
		"saldo": gin.H{
			"total":        customer.Balance,
			"data_extrato": time.Now(),
			"limite":       customer.Limit,
		},
		"ultimas_transacoes": transactions,
	}

	c.JSON(http.StatusOK, response)
}
