package main

import (
	"blood-donation-backend/bloodinfo"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"os"
)

func main() {
	var err error

	// Get PostgreSQL connection details from environment variables
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbName := os.Getenv("DB_NAME")
	dbPassword := os.Getenv("DB_PASSWORD")

	// Construct the connection string
	connectionString := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=disable", dbHost, dbPort, dbUser, dbName, dbPassword)

	// Connect to the PostgreSQL database
	db, err := gorm.Open(postgres.Open(connectionString), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	// AutoMigrate will create the tables based on the struct definitions
	err = db.AutoMigrate(&bloodinfo.User{})
	if err != nil {
		log.Fatal(err)
	}
	err = db.AutoMigrate(&bloodinfo.Station{})
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Server is running on port 8080")
	// log.Fatal(http.ListenAndServe(":8080", router))
}
