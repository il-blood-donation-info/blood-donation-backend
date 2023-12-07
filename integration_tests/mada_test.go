package integration_tests

import (
	"bytes"
	"context"
	"fmt"
	"github.com/agiledragon/gomonkey/v2"
	"github.com/il-blood-donation-info/blood-donation-backend/pkg/api"
	"github.com/il-blood-donation-info/blood-donation-backend/pkg/scraper"
	"github.com/il-blood-donation-info/blood-donation-backend/server"
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

func setupServer() *server.StrictBloodInfoServer {
	db := setupDatabase()
	srv := server.NewStrictBloodInfoServer(db)
	return &srv
}

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

	err = db.AutoMigrate(&api.User{}, &api.Station{}, &api.StationStatus{}, &api.StationSchedule{})
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
	err := db.Migrator().DropTable(&api.User{}, &api.Station{}, &api.StationStatus{}, &api.StationSchedule{})
	if err != nil {
		log.Fatal(err)
	}
	scraper.CloseDbConnection(db)
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
	patches := gomonkey.ApplyFunc(time.Now, func() time.Time {
		return time.Unix(1697619600, 0)
	})
	defer patches.Reset()

	srv := setupServer()
	defer teardown(srv.Db)
	ResetMocks()

	MockClient.DoFunc = func(*http.Request) (*http.Response, error) {
		dir, err := filepath.Abs(filepath.Dir("."))
		if err != nil {
			log.Fatal(err)
		}

		filePath := filepath.Join(dir, "mada_data.json")

		mockResponse, err := os.ReadFile(filePath)
		if err != nil {
			log.Fatal(err)
		}

		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewBufferString(string(mockResponse))),
		}, nil
	}
	s := scraper.Scraper{
		Client: MockClient,
	}
	madaResponse, err := s.ScrapeMada()
	if err != nil {
		t.Fatalf("Failed to scrape Mada: %s", err)
	}

	// Basic check if we got some data
	if len(madaResponse) == 0 {
		t.Fatal("Received empty response from Mada")
	}
	today := time.Date(2023, 10, 18, 0, 0, 0, 0, time.UTC)
	oneDayBefore := today.AddDate(0, 0, -1)
	log.Println("SaveData")
	p := scraper.ScheduleDataWriter{
		DB:        srv.Db,
		SinceTime: today,
	}
	err = p.SaveData(madaResponse)
	if err != nil {
		log.Fatalf("Failed to SaveData: %s", err)
	}

	scheduler := server.NewScheduler(server.WithSinceDate(today))
	schedule, err := scheduler.GetStationsFullSchedule(srv.Db)
	if err != nil {
		t.Fatal(err)
	}

	scheduledYesterday := schedule.FilterByDate(oneDayBefore)
	if len(scheduledYesterday) > 0 {
		t.Fatal("no yesterday dates should be present in schedule")
	}
	todaySchedule := schedule.FilterByDate(today)
	if len(todaySchedule) != 5 {
		t.Fatal(fmt.Sprintf("today should have 5 schedule points, has: %d", len(todaySchedule)))
	}

	// Check current status of station in schedule
	// in todaySchedule find the station with name "TestName"
	// check that it is open
	for _, sc := range todaySchedule {
		if sc.StationName == "TestName" {
			if sc.LastStatus != true {
				t.Fatal("TestName should be open")
			}
		}
	}

	// find the right station
	stationName := "TestName"
	station := api.Station{}
	srv.Db.Find(&station, api.Station{Name: stationName})
	if station.Id == 0 {
		t.Fatal("station not found")
	}

	// make a put call to station
	updateRequest := api.UpdateStationRequestObject{
		Id:   station.Id,
		Body: &api.UpdateStationJSONRequestBody{IsOpen: false},
	}
	_, err = srv.UpdateStation(context.TODO(), updateRequest)
	if err != nil {
		t.Fatal(err)
	}

	schedule, err = scheduler.GetStationsFullSchedule(srv.Db)
	if err != nil {
		t.Fatal(err)
	}
	todaySchedule = schedule.FilterByDate(today)

	for _, sc := range todaySchedule {
		if sc.StationName == "TestName" {
			if sc.LastStatus != false {
				t.Fatal("TestName should be closed")
			}
		}
	}
}
