// scraper/mada.go

package scraper

import (
	"blood-donation-backend/bloodinfo"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

const customDateLayout = "2006-01-02"
const customDateTimeLayout = "2006-01-02 15:04"

var dbManager *DBManager

type DBManager struct {
	DB *gorm.DB
}

type DonationDetail struct {
	Name         string    `json:"Name"`
	City         string    `json:"City"`
	Street       string    `json:"Street"`
	NumHouse     string    `json:"NumHouse"`
	FromHour     string    `json:"FromHour"`
	ToHour       string    `json:"ToHour"`
	DateDonation time.Time `json:"DateDonation"`
}

type MadaResponse struct {
	ErrorCode string `json:"ErrorCode"`
	ErrorMsg  string `json:"ErrorMsg"`
	Result    string `json:"Result"`
}

var Client Doer = &http.Client{}

type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

func init() {
    // This will ensure that initDb is called only once
	log.Println("Call init function in mada file")
	initDb()
}

func (d *DonationDetail) UnmarshalJSON(data []byte) error {
	type Alias DonationDetail
	aux := &struct {
		DateStr string `json:"DateDonation"`
		*Alias
	}{
		Alias: (*Alias)(d),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if aux.DateStr != "" {
		t, err := time.Parse(customDateLayout, aux.DateStr[:10]) // Taking only the date part
		if err != nil {
			return err
		}
		d.DateDonation = t
	}
	return nil
}

func ScrapeMada() ([]DonationDetail, error) {
	payload := map[string]interface{}{
		"RequestHeader": map[string]string{
			"Application": "101",
			"Module":      "BloodBank",
			"Function":    "GetAllDetailsDonations",
			"Token":       "",
		},
		"RequestData": "",
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %v", err)
	}

	// Create a new request
	req, err := http.NewRequest("POST", "https://www.mdais.org/umbraco/api/invoker/execute", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Add the headers
	req.Header.Set("authority", "www.mdais.org")
	req.Header.Set("accept", "application/json, text/plain, */*")
	req.Header.Set("accept-language", "he")
	req.Header.Set("cache-control", "no-cache")
	req.Header.Set("content-type", "application/json")
	req.Header.Set("pragma", "no-cache")
	req.Header.Set("referer", "https://www.mdais.org/blood-donation")

	// Execute the request
	resp, err := Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data from Mada: %v", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Printf("failed to close reader body: %v", err)
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	var response MadaResponse
	if err = json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("error parsing JSON response: %v", err)
	}

	var donationDetails []DonationDetail
	if err = json.Unmarshal([]byte(response.Result), &donationDetails); err != nil {
		return nil, fmt.Errorf("error parsing Result string into donation details: %v", err)
	}

	return donationDetails, nil
}

func initDb()error{
    log.Println("Initializing the database...")

    if dbManager  != nil{
        log.Println("Db already initialised, exiting initDb")
        return nil
    }

	//Remove duplicate connection-DB code, new file for main server and scraper?
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
        return err
	}

	err = db.AutoMigrate(&bloodinfo.Station{})
	if err != nil {
        log.Println("Error during database migration:", err)
		log.Fatal(err)
        return err
	}

    dbManager = &DBManager{DB: db}

    log.Println("Database successfully initialised")
    return nil
}

func closeDbConnection() {
    // Close the database connection
    if dbManager != nil {
        sqlDB, err := dbManager.DB.DB()
        if err != nil {
            log.Fatal(err)
        }
        sqlDB.Close()
        log.Println("Database connection closed")
    }
}

func ConvertDonationToStation(d DonationDetail) bloodinfo.Station {
    log.Println("Convert  data for " + d.Name)
	existingStation, err := findStationByName(d.Name)

	if err != nil {
		// Handle the error
		log.Fatal(err)
	}

	stationSchedule := bloodinfo.StationSchedule{
		Date:      d.DateDonation,
		OpenTime:  d.FromHour,
		CloseTime: d.ToHour,
	}

	// If the station exists, add the new schedule to its StationSchedules
	if existingStation != nil {
	    log.Println("Station already exist %", existingStation.Id)

        // Check if StationSchedule is nil, and initialize it with an empty slice if necessary
        if existingStation.StationSchedule == nil {
            existingStation.StationSchedule = &[]bloodinfo.StationSchedule{}
        }

        // Append the new schedule to StationSchedules
        *existingStation.StationSchedule = append(*existingStation.StationSchedule, stationSchedule)

		return *existingStation
	}

    log.Println("Station does not exist, need to create it")
	// If the station doesn't exist, create a new one
	station := bloodinfo.Station{
		Address:         fmt.Sprintf("%s, %s %s", d.Street, d.NumHouse, d.City), //need formatting
		Name:            d.Name,                                                 //need formatting (at least trim)
		StationSchedule: &[]bloodinfo.StationSchedule{stationSchedule},
		StationStatus:   &[]bloodinfo.StationStatus{},
	}

	return station
}

func SaveData(donationDetails []DonationDetail) error {
	//todo : Clean DB before adding ? Filter for adding just stations for today ?
    log.Println("Beginning of SaveData")

	for _, donation := range donationDetails {
		station := ConvertDonationToStation(donation)

	    log.Println("Ready to handle stationData " + station.Name)
		result := dbManager.DB.Create(&station)
		if result.Error != nil {
			log.Fatal(result.Error)
		}
	}

	fmt.Println("Bulk insert completed successfully.")
	return nil
}

func findStationByName(name string) (*bloodinfo.Station, error) {
    log.Println("Finding station by name:", name)  // Add this line

	var station bloodinfo.Station
	result := dbManager.DB.Where("name = ?", name).First(&station)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// Station not found
			return nil, nil
		}
		// Handle other errors
		log.Fatal(result.Error)
		return nil, result.Error
	}
	return &station, nil
}
