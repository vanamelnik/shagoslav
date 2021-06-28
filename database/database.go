package database

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

const (
	ErrNotFound dbError = "database: resource not found"
)

type dbError string

func (e dbError) Error() string {
	return string(e)
}

// type Database interface {
// 	AutoMigrate() error
// 	DestructiveReset() error
// }

func NewDB() *sql.DB {
	connStr := "host=localhost port=5432 user=postgres password=qwe123 dbname=shagoslav_dev sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("database service: %v", err)
	}
	return db
}

// type DB struct {
// 	db *sql.DB
// }

// func (dbs *DB) Close() {
// 	dbs.db.Close()
// }
