package database

import (
	"database/sql"
	"log"
)

var db *sql.DB

func connect() *sql.DB {
	var err error
	db, err = sql.Open("mysql", "root:q1w2r4e3@tcp(localhost:3306)/rinha")
	if err != nil {
		log.Fatal(err)
	}

	return db
}
