package main

import (
	"database/sql"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

type Customer struct {
	ID      int `json:"id"`
	Balance int `json:"balance"`
	Limit   int `json:"limit"`
}

type TransactionPayload struct {
	Value       int    `json:"valor" binding:"required"`
	Type        string `json:"tipo" binding:"required,oneof=c d"`
	Description string `json:"descricao"`
}

type Transaction struct {
	Value       int       `json:"value"`
	Type        string    `json:"type"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

var (
	db *sql.DB
	mu sync.Mutex
)

func initDB() *sql.DB {
	var err error
	db, err = sql.Open("mysql", "root:root@tcp(mysql:3306)/rinha?parseTime=true")
	if err != nil {
		log.Fatal(err)
	}

	return db
}

func main() {
	r := gin.Default()

	initDB()

	r.GET("/clientes/:id/extrato", listBankStatement)
	r.POST("/clientes/:id/transacoes", createTransaction)

	if err := r.Run(":8000"); err != nil {
		log.Fatal(err)
	}
}

func createTransaction(c *gin.Context) {
	id := c.Param("id")

	var payload TransactionPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	if payload.Type != "c" && payload.Type != "d" {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Tipo de transação inválido"})
		return
	}

	mu.Lock()
	defer mu.Unlock()

	customerQuery := "SELECT balance, `limit` FROM customers WHERE id = ?"
	var customer Customer

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
		"saldo":  newBalance,
	}

	c.JSON(http.StatusOK, response)
}

func listBankStatement(c *gin.Context) {
	id := c.Param("id")

	query := `
		SELECT value, type, description, transactions.created_at
		FROM transactions
		INNER JOIN customers ON transactions.customer_id = customers.id
		WHERE customers.id = ?
		ORDER BY transactions.created_at DESC
		LIMIT 10
	`

	rows, err := db.Query(query, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	defer rows.Close()

	var transactions []Transaction = []Transaction{}

	for rows.Next() {
		var transaction Transaction
		if err := rows.Scan(&transaction.Value, &transaction.Type, &transaction.Description, &transaction.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		transactions = append(transactions, transaction)
	}

	customerQuery := "SELECT balance, `limit` FROM customers WHERE id = ?"
	var customer Customer
	err = db.QueryRow(customerQuery, id).Scan(&customer.Balance, &customer.Limit)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Erro ao obter informações do cliente"})
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
