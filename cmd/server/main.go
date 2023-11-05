package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/il-blood-donation-info/blood-donation-backend/pkg/api"
	"github.com/il-blood-donation-info/blood-donation-backend/server"
	middleware "github.com/oapi-codegen/nethttp-middleware"
	"github.com/rs/cors"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"net"
	"net/http"
	"os"
)

func main() {
	var err error
	certFile := flag.String("certfile", "cert.pem", "certificate PEM file")
	keyFile := flag.String("keyfile", "key.pem", "key PEM file")
	flag.Parse()

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
	err = db.AutoMigrate(&api.User{}, &api.Station{}, &api.StationStatus{}, &api.StationSchedule{})
	if err != nil {
		log.Fatal(err)
	}

	swagger, err := api.GetSwagger()
	if err != nil {
		log.Fatalf("Error loading swagger spec\n: %s", err)
	}
	swagger.Servers = nil

	strictBloodInfoServer := server.NewStrictBloodInfoServer(db)

	// We can add 2 kinds of middlewares:
	// - "strict" middlewares here deal in per-request generated `api` types.
	// - r.Use() chi middlewares below deal in de-facto standard `http.Handler`.
	strictMiddlewares := []api.StrictMiddlewareFunc{}
	strictHandler := api.NewStrictHandler(strictBloodInfoServer, strictMiddlewares)

	r := chi.NewRouter()
	r.Use(cors.Default().Handler) // Tell browsers cross-origin requests OK from any domain.
	r.Use(middleware.OapiRequestValidator(swagger))
	api.HandlerFromMux(strictHandler, r)
	s := &http.Server{
		Handler: r,
		Addr:    net.JoinHostPort("0.0.0.0", "8443"),
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS13,
		},
	}
	log.Println("Server is running on port 8443")
	log.Fatal(s.ListenAndServeTLS(*certFile, *keyFile))
}
