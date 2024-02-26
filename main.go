package main

import (
	"log"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	r := gin.Default()

	r.GET("/clientes/:id/extrato", handler.listBankStatement)
	r.POST("/clientes/:id/transacoes", handler.createTransaction)

	if err := r.Run(":9999"); err != nil {
		log.Fatal(err)
	}
}
