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
	"strings"
	"time"
)

const customDateLayout = "2006-01-02"

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
	//This will ensure that initDb is called only once
	log.Println("Call init function in mada file")

	err := initDb()
	if err != nil {
		log.Fatal(err)
	}
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

func initDb() error {
	log.Println("Initializing the database...")

	if dbManager != nil {
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

	fmt.Printf("host=%s port=%s user=%s dbname=%s password=%s sslmode=disable", dbHost, dbPort, dbUser, dbName, dbPassword)

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

	err = db.AutoMigrate(&bloodinfo.StationStatus{})
	if err != nil {
		log.Println("Error during database migration:", err)
		log.Fatal(err)
		return err
	}

	err = db.AutoMigrate(&bloodinfo.StationSchedule{})
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
		err = sqlDB.Close()
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Database connection closed")
	}
}

func ConvertDonationToStation(d DonationDetail) bloodinfo.Station {
	stationName := strings.TrimSpace(d.Name)
	stationAddress := strings.TrimSpace(fmt.Sprintf("%s %s %s", strings.TrimSpace(d.City), strings.TrimSpace(d.NumHouse), strings.TrimSpace(d.Street)))

	//log.Println("Convert  data for " + stationName)
	existingStation, err := findStationByName(stationName)

	if err != nil {
		// Handle the error
		log.Fatal(err)
	}

	stationSchedule := bloodinfo.StationSchedule{
		Date:      d.DateDonation,
		OpenTime:  d.FromHour,
		CloseTime: d.ToHour,
	}

	today := time.Now()
	var stationScheduleIsPertinent = stationSchedule.Date.Year() > today.Year() || (stationSchedule.Date.Year() == today.Year() && stationSchedule.Date.YearDay() >= today.YearDay())

	// If the station exists, add the new schedule to its StationSchedules
	if existingStation != nil {
		//log.Println("Station already exist ", existingStation.Id)

		// Check if StationSchedule is nil, and initialize it with an empty slice if necessary
		if stationScheduleIsPertinent {
			if existingStation.StationSchedule == nil {
				existingStation.StationSchedule = &[]bloodinfo.StationSchedule{}
			}
			// Append the new schedule to StationSchedules
			*existingStation.StationSchedule = append(*existingStation.StationSchedule, stationSchedule)
		}

		return *existingStation
	}

	log.Println("Station does not exist, need to create it")
	// If the station doesn't exist, create a new one
	station := bloodinfo.Station{
		Address: stationAddress,
		Name:    stationName,
	}
	if stationScheduleIsPertinent {
		station.StationSchedule = &[]bloodinfo.StationSchedule{stationSchedule}
	}

	return station
}

func SaveData(donationDetails []DonationDetail) error {
	//todo : Clean DB before adding ? Filter for adding just stations for today ?
	log.Println("Beginning of SaveData")

	var resultStations []bloodinfo.Station

	for _, donation := range donationDetails {
		station := ConvertDonationToStation(donation)

		existingIndex := findStationIndexByName(resultStations, station.Name)

		if existingIndex != -1 {
			// Concatenate the StationSchedules for stations with the same name
			resultStations[existingIndex].StationSchedule = concatenateSchedules(resultStations[existingIndex].StationSchedule, station.StationSchedule)
		} else {
			resultStations = append(resultStations, station)
		}
	}

	//For testing purpose: All the stations & all associated scheduled are created
	for _, station := range resultStations {
		tx := dbManager.DB.Begin()

		//log.Printf("station: %+v", station)
		//log.Println("Handling " + station.Name)

		// Handle station schedules: it's already >= today
		for i := range *station.StationSchedule {
			schedule := &(*station.StationSchedule)[i]

			//log.Println("Checking schedule ", schedule.Date, schedule.OpenTime, schedule.CloseTime)

			// Check if the schedule exists in the database
			schedule.StationId = station.Id
			existingSchedule, err := findSchedule(*schedule)
			if err != nil {
				tx.Rollback()
				return err
			}

			if existingSchedule == nil {
				if isToday(*schedule) {
					//log.Println("Schedule not existing, and is today: is_open = true")
					stationStatus := bloodinfo.StationStatus{
						IsOpen:    true,
						CreatedAt: time.Now(),
					}
					schedule.StationStatus = &[]bloodinfo.StationStatus{stationStatus}
					//log.Printf("schedule: %+v", schedule)
				}
			} else {
				schedule.Id = existingSchedule.Id
			}
		}

		//log.Printf("station: %+v", station)
		//log.Printf("stationSchedule: %+v", station.StationSchedule)
		//for _, schedule := range *station.StationSchedule {
		//log.Printf("schedule: %+v", schedule)
		//if schedule.StationStatus != nil {
		//for _, status := range *schedule.StationStatus {
		//log.Printf("status: %+v", status)
		//}
		//}
		//}
		if station.Id == 0 {
			// Station does not exist, create it
			//log.Println("Create")
			if err := tx.Create(&station).Error; err != nil {
				tx.Rollback()
				return err
			}
		} else {
			// Station already exists, update it
			//log.Println("Update")
			if err := tx.Save(&station).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
		tx.Commit()

		//log.Printf("station: %+v", station)
		//log.Printf("stationSchedule: %+v", station.StationSchedule)
		//for _, schedule := range *station.StationSchedule {
		//	//log.Printf("schedule: %+v", schedule)
		//	if schedule.StationStatus != nil {
		//		for _, status := range *schedule.StationStatus {
		//			//log.Printf("status: %+v", status)
		//		}
		//	}
		//}

	}

	fmt.Println("Bulk insert completed successfully.")

	return nil
}

func findStationByName(name string) (*bloodinfo.Station, error) {
	//log.Println("Finding station by name:", name) // Add this line

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

// findStationIndexByName finds the index of a station with the given name in the list
func findStationIndexByName(stations []bloodinfo.Station, name string) int {
	for i, station := range stations {
		if station.Name == name {
			return i
		}
	}
	return -1
}

func findSchedule(schedule bloodinfo.StationSchedule) (*bloodinfo.StationSchedule, error) {
	var foundSchedule bloodinfo.StationSchedule

	result := dbManager.DB.Where("date = ? AND open_time = ? AND close_time = ? AND station_id = ?",
		schedule.Date, schedule.OpenTime, schedule.CloseTime, schedule.StationId).
		First(&foundSchedule)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// Schedule not found
			return nil, nil
		}
		return nil, result.Error
	}

	return &foundSchedule, nil
}

// concatenateSchedules concatenates two StationSchedule arrays
func concatenateSchedules(schedule1 *[]bloodinfo.StationSchedule, schedule2 *[]bloodinfo.StationSchedule) *[]bloodinfo.StationSchedule {
	if schedule1 == nil {
		return schedule2
	}
	if schedule2 == nil {
		return schedule1
	}

	concatenated := append(*schedule1, *schedule2...)
	return &concatenated
}

func isToday(schedule bloodinfo.StationSchedule) bool {
	// Get the current date
	currentDate := time.Now().UTC().Truncate(24 * time.Hour)

	// Truncate the time part from the schedule date
	scheduleDate := schedule.Date.UTC().Truncate(24 * time.Hour)

	// Compare the truncated dates
	return scheduleDate.Equal(currentDate)
}
