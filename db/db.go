package db

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var dbConnection *sql.DB

func GetDBConnection() *sql.DB {

	connectionString := os.Getenv("PRONGHORN_DATABASE_CONNECTION_STRING")

	if dbConnection != nil {
		return dbConnection
	}

	db, err := sql.Open("postgres", connectionString)

	if err != nil {
		log.Fatal("Failed to connect to database", err)
	}

	dbConnection = db

	return dbConnection
}
