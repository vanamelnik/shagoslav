package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

const (
	// ErrNotFound returns when the entry in db is not found
	ErrNotFound dbError = "database: resource not found"
)

type dbError string

func (e dbError) Error() string {
	return string(e)
}

// NewDB connects to Postgresql database
// It panics if there's any error
func NewDB() *sql.DB {
	connStr := "host=localhost port=5432 user=postgres password=qwe123 dbname=shagoslav_dev sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("database service: %v", err)
	}
	fmt.Println("successfully connected to Postgres database shagoslav_dev")
	return db
}

// type DB struct {
// 	db *sql.DB
// }

// func (dbs *DB) Close() {
// 	dbs.db.Close()
// }
