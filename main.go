package main

import (
	"blood-donation-backend/api"
	"blood-donation-backend/server"
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	middleware "github.com/oapi-codegen/nethttp-middleware"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"net"
	"net/http"
	"os"
)

var tokenAuth *jwtauth.JWTAuth

func init() {
	tokenAuth = jwtauth.New("HS256", []byte("secret"), nil)

	// For debugging/example purposes, we generate and print
	// a sample jwt token with claims `user_id:123` here:
	_, tokenString, _ := tokenAuth.Encode(map[string]interface{}{"user_id": 123})
	fmt.Printf("DEBUG: a sample jwt is %s\n\n", tokenString)
}

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
	err = db.AutoMigrate(&api.User{})
	if err != nil {
		log.Fatal(err)
	}
	err = db.AutoMigrate(&api.Station{})
	if err != nil {
		log.Fatal(err)
	}

	swagger, err := api.GetSwagger()
	if err != nil {
		log.Fatalf("Error loading swagger spec\n: %s", err)
	}
	swagger.Servers = nil

	strictBloodInfoServer := server.NewStrictBloodInfoServer(db)
	strictHandler := api.NewStrictHandler(strictBloodInfoServer, nil)

	r := chi.NewRouter()
	// Seek, verify and validate JWT tokens
	r.Use(jwtauth.Verifier(tokenAuth))

	// Handle valid / invalid tokens. In this example, we use
	// the provided authenticator middleware, but you can write your
	// own very easily, look at the Authenticator method in jwtauth.go
	// and tweak it, its not scary.
	r.Use(jwtauth.Authenticator)

	validatorOptions := &middleware.Options{}

	validatorOptions.Options.AuthenticationFunc = func(c context.Context, input *openapi3filter.AuthenticationInput) error {
		fmt.Println(">>>> INSIDE AuthenticationFunc")
		return nil
	}
	r.Use(middleware.OapiRequestValidatorWithOptions(swagger, validatorOptions))

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
