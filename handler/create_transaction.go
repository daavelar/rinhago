package handlers

import (
	"net/http"
	"rinhago/structs"
	"time"

	"github.com/gin-gonic/gin"
)

func createTransaction(c *gin.Context) {
	id := c.Param("id")

	var payload structs.TransactionPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	if payload.Type != "c" && payload.Type != "d" {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Tipo de transação inválido"})
		return
	}

	customerQuery := "SELECT balance, `limit` FROM customers WHERE id = ?"
	var customer structs.Customer
	db := database.connect()

	err := db.QueryRow(customerQuery, id).Scan(&customer.Balance, &customer.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	var newBalance int = 0

	if payload.Type == "c" {
		newBalance = payload.Value + customer.Balance
	}

	if payload.Type == "d" {
		newBalance = customer.Balance - payload.Value
	}

	if payload.Type == "d" && newBalance*-1 > customer.Limit {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Débito ultrapassa o limite do cliente"})
		return
	}

	tx, err := db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao iniciar a transação"})
		return
	}

	defer func() {
		if err := recover(); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao salvar a transação no banco de dados"})
		}
	}()

	insertQuery := "INSERT INTO transactions (customer_id, value, type, description, created_at) VALUES (?, ?, ?, ?, ?)"
	_, err = tx.Exec(insertQuery, id, payload.Value, payload.Type, payload.Description, time.Now())
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao salvar a transação no banco de dados"})
		return
	}

	updateBalanceQuery := "UPDATE customers SET balance = ? WHERE id = ?"
	_, err = tx.Exec(updateBalanceQuery, newBalance, id)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao finalizar a transação"})
		return
	}

	response := gin.H{
		"limite": customer.Limit,
		"saldo":  customer.Balance,
	}

	c.JSON(http.StatusOK, response)
}
