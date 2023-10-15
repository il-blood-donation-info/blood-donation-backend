package main

import (
	"blood-donation-backend/bloodinfo"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"log"
	"net/http"
	"os"
)

var (
	db *gorm.DB
)

func main() {
	// Connect to the PostgreSQL database
	var err error
	// Get PostgreSQL connection details from environment variables
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbName := os.Getenv("DB_NAME")
	dbPassword := os.Getenv("DB_PASSWORD")

	// Construct the connection string
	connectionString := fmt.Sprintf("host=%s port=%s user.go=%s dbname=%s password=%s sslmode=disable", dbHost, dbPort, dbUser, dbName, dbPassword)

	// Connect to the PostgreSQL database
	db, err = gorm.Open("postgres", connectionString)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// AutoMigrate will create the tables based on the struct definitions
	db.AutoMigrate(&bloodinfo.User{})
	db.AutoMigrate(&bloodinfo.Station{})

	router := mux.NewRouter()
	router.HandleFunc("/stations", GetStations).Methods("GET")
	router.HandleFunc("/stations/{id}", UpdateStation).Methods("PUT")
	router.HandleFunc("/users", GetUsers).Methods("GET")
	router.HandleFunc("/users", CreateUser).Methods("POST")
	router.HandleFunc("/users/{id}", UpdateUser).Methods("PUT")
	router.HandleFunc("/users/{id}", DeleteUser).Methods("DELETE")

	log.Println("Server is running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}

func GetStations(w http.ResponseWriter, r *http.Request) {
	// Logic to get all stations from the database and return as JSON
}

func UpdateStation(w http.ResponseWriter, r *http.Request) {
	// Logic to update a station's current status (isOpen) based on the provided ID
}

func GetUsers(w http.ResponseWriter, r *http.Request) {
	// Logic to get all users from the database and return as JSON
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	// Logic to create a new user.go based on the provided JSON data
}

func UpdateUser(w http.ResponseWriter, r *http.Request) {
	// Logic to update a user.go based on the provided ID and JSON data
}

func DeleteUser(w http.ResponseWriter, r *http.Request) {
	// Logic to "soft" delete a user.go based on the provided ID
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "JSON marshaling failed")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
