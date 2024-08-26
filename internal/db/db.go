package db

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var DB *sqlx.DB

func InitDB() {
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	dataSourceName := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	maxRetries := 5
	retryInterval := 2 * time.Second

	var err error
	for i := 0; i < maxRetries; i++ {
		DB, err = sqlx.Connect("postgres", dataSourceName)
		if err == nil {
			log.Println("Database connection established")
			return
		}

		log.Printf("Error connecting to the database (attempt %d/%d): %v", i+1, maxRetries, err)
		time.Sleep(retryInterval)
	}

	log.Fatalf("Failed to connect to the database after %d attempts: %v", maxRetries, err)
}
