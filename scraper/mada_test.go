package scraper

import (
	"blood-donation-backend/bloodinfo"
	"bytes"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

import (
	"net/http"
)

var setupDone sync.Once

// Mock HTTP client
type MockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

var (
	// MockClient is an instantiation of MockHTTPClient with DoFunc set
	MockClient = &MockHTTPClient{}
)

func setupDatabase() *gorm.DB {
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbName := os.Getenv("DB_NAME")
	dbPassword := os.Getenv("DB_PASSWORD")
	connectionString := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=disable", dbHost, dbPort, dbUser, dbName, dbPassword)
	db, err := gorm.Open(postgres.Open(connectionString), &gorm.Config{})
	if err != nil {
		panic("failed to connect to test database")
	}
	return db
}

func SetupTests() {
	db := setupDatabase()
	// AutoMigrate will create the tables based on the struct definitions
	err := db.AutoMigrate(&bloodinfo.User{})
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
}

// Do is the mock client's `Do` function
func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.DoFunc(req)
}

// ResetMocks resets all mocks (useful in tests to ensure clean state)
func ResetMocks() {
	MockClient = &MockHTTPClient{}
}

func TestScrapeMada(t *testing.T) {
	setupDone.Do(SetupTests)
	ResetMocks()

	MockClient.DoFunc = func(*http.Request) (*http.Response, error) {
		dir, err := filepath.Abs(filepath.Dir("."))
		if err != nil {
			log.Fatal(err)
		}
		fileName := "mada_test_data.json"
		filePath := filepath.Join(dir, fileName)

		mockResponse, err := os.ReadFile(filePath)
		if err != nil {
			log.Fatal(err)
		}

		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewBufferString(string(mockResponse))),
		}, nil
	}
	madaResponse, err := ScrapeMada()
	if err != nil {
		t.Fatalf("Failed to scrape Mada: %s", err)
	}

	for _, result := range madaResponse {
		fmt.Printf("Name: %s\n", result.Name)
		fmt.Printf("Address: %s %s %s\n", result.City, result.Street, result.NumHouse)
		fmt.Printf("Open Time: %s\n", result.FromHour)
		fmt.Printf("Close Time: %s\n\n", result.ToHour)
		fmt.Printf("Datetime : %s\n\n", result.DateDonation)
	}

	// Basic check if we got some data
	if len(madaResponse) == 0 {
		t.Fatal("Received empty response from Mada")
	}

	SaveData(madaResponse)
	schedule, err := bloodinfo.GetStationsFullSchedule(dbManager.DB)
	if err != nil {
		t.Fatalf("Failed to get schedule Mada: %s", err)
	}
	fmt.Println(schedule)
}

func TestScraper2DB(t *testing.T) {
	defer closeDbConnection()
	setupDone.Do(SetupTests)
	madaResponse, err := ScrapeMada()
	if err != nil {
		t.Fatalf("Failed to scrape Mada: %s", err)
	}

	// 	for _, result := range madaResponse {
	// 		fmt.Printf("Name: %s\n", result.Name)
	// 		fmt.Printf("Address: %s %s %s\n", result.City, result.Street, result.NumHouse)
	// 		fmt.Printf("Open Time: %s\n", result.FromHour)
	// 		fmt.Printf("Close Time: %s\n\n", result.ToHour)
	// 		fmt.Printf("Datetime : %s\n\n", result.DateDonation)
	// 	}

	// Basic check if we got some data
	if len(madaResponse) == 0 {
		t.Fatal("Received empty response from Mada")
	}

	log.Println("SaveData")
	SaveData(madaResponse)
}
