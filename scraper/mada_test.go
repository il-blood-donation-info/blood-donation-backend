package main

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
	"testing"
	"time"
)

import (
	"net/http"
)

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
	dbUser := os.Getenv("DB_USER_TEST")
	dbName := os.Getenv("DB_NAME_TEST")
	dbPassword := os.Getenv("DB_PASSWORD")
	connectionString := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=disable", dbHost, dbPort, dbUser, dbName, dbPassword)

	db, err := gorm.Open(postgres.Open(connectionString), &gorm.Config{})
	if err != nil {
		panic("failed to connect to test database")
	}
	err = db.AutoMigrate(&bloodinfo.User{}, &bloodinfo.Station{}, &bloodinfo.StationStatus{}, &bloodinfo.StationSchedule{})
	if err != nil {
		log.Fatalf("Failed to migrate... %+v", err)
	}
	return db
}

func teardown(db *gorm.DB) {
	if db == nil {
		return
	}
	// Perform teardown tasks here
	err := db.Migrator().DropTable(&bloodinfo.User{}, &bloodinfo.Station{}, &bloodinfo.StationStatus{}, &bloodinfo.StationSchedule{})
	if err != nil {
		log.Fatal(err)
	}
	closeDbConnection(db)
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
	db := setupDatabase()
	defer teardown(db)
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
		fmt.Printf("Close Time: %s\n", result.ToHour)
		fmt.Printf("Datetime : %s\n\n", result.DateDonation)
	}

	// Basic check if we got some data
	if len(madaResponse) == 0 {
		t.Fatal("Received empty response from Mada")
	}

	log.Println("SaveData")
	p := DataWriter{
		DB: db,
	}
	err = p.SaveData(madaResponse)
	if err != nil {
		log.Fatalf("Failed to SaveData: %s", err)
	}

	//after
	fmt.Println(">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>")
	today := time.Date(2023, 10, 18, 0, 0, 0, 0, time.UTC)
	oneDayBefore := today.AddDate(0, 0, -1)
	s := bloodinfo.NewScheduler(bloodinfo.WithSinceDate(today))
	schedule, err := s.GetStationsFullSchedule(db)
	if err != nil {
		t.Fatal(err)
	}

	scheduledYesterday := schedule.FilterByDate(oneDayBefore)
	if len(scheduledYesterday) > 0 {
		t.Fatal("no yesterday dates should be present in schedule")
	}

	todaySchedule := schedule.FilterByDate(today)
	if len(todaySchedule) != 8 {
		t.Fatal(fmt.Sprintf("today should have ## schedule points, has: %d", len(todaySchedule)))
	}
}
