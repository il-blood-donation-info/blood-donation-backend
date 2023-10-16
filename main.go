package main

import (
	"blood-donation-backend/bloodinfo"
	"fmt"
	"github.com/go-chi/chi/v5"
	middleware "github.com/oapi-codegen/nethttp-middleware"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"net"
	"net/http"
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
	err = db.AutoMigrate(&bloodinfo.StationStatus{})
	if err != nil {
		log.Fatal(err)
	}
	err = db.AutoMigrate(&bloodinfo.StationSchedule{})
	if err != nil {
		log.Fatal(err)
	}

	swagger, err := bloodinfo.GetSwagger()
	if err != nil {
		log.Fatalf("Error loading swagger spec\n: %s", err)
	}
	swagger.Servers = nil

	strictBloodInfoServer := bloodinfo.NewStrictBloodInfoServer(db)
	strictHandler := bloodinfo.NewStrictHandler(strictBloodInfoServer, nil)

	r := chi.NewRouter()
	r.Use(middleware.OapiRequestValidator(swagger))
	bloodinfo.HandlerFromMux(strictHandler, r)
	s := &http.Server{
		Handler: r,
		Addr:    net.JoinHostPort("0.0.0.0", "8080"),
	}
	log.Println("Server is running on port 8080")
	log.Fatal(s.ListenAndServe())
}
