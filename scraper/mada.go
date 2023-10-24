// scraper/mada.go

package scraper

import (
	"blood-donation-backend/bloodinfo"
	"bytes"
	"encoding/json"
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
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbName := os.Getenv("DB_NAME")
	dbPassword := os.Getenv("DB_PASSWORD")

	connectionString := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=disable", dbHost, dbPort, dbUser, dbName, dbPassword)

	// Connect to the PostgreSQL database
	db, err := gorm.Open(postgres.Open(connectionString), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
		return err
	}

	err = db.AutoMigrate(&bloodinfo.Station{}, &bloodinfo.StationStatus{}, &bloodinfo.StationSchedule{})
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

func SaveData(donationDetails []DonationDetail) error {
	tx := dbManager.DB.Begin()
	if tx.Error != nil {
		return tx.Error // Check for an error when starting the transaction
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback() // Rollback the transaction in case of a panic
		}
	}()

	err := processDonationDetails(tx, donationDetails)
	if err != nil {
		tx.Rollback() // Rollback the transaction if an error occurs
		return err
	}

	tx.Commit() // Commit the transaction if no errors occurred
	return nil
}

func processDonationDetails(tx *gorm.DB, donationDetails []DonationDetail) error {
	var schedulesIds []int64
	for _, donation := range donationDetails {
		stationName := strings.TrimSpace(donation.Name)
		stationAddress := strings.TrimSpace(fmt.Sprintf("%s %s %s", strings.TrimSpace(donation.City), strings.TrimSpace(donation.NumHouse), strings.TrimSpace(donation.Street)))

		var station = bloodinfo.Station{}
		if err := tx.FirstOrInit(&station, bloodinfo.Station{Name: stationName}).Error; err != nil {
			log.Printf("Error while fetching or initializing station: %v", err)
			return err
		}
		station.Address = stationAddress

		if !isDatePassed(donation.DateDonation) {
			var schedule = bloodinfo.StationSchedule{}
			if err := tx.FirstOrInit(&schedule, bloodinfo.StationSchedule{
				Date:      donation.DateDonation,
				OpenTime:  donation.FromHour,
				CloseTime: donation.ToHour,
			}).Error; err != nil {
				log.Printf("Error while fetching or initializing schedule: %v", err)
				return err
			}

			if schedule.Id == nil && isScheduleToday(schedule) {
				schedule.StationStatus = &[]bloodinfo.StationStatus{{IsOpen: true}}
			}

			if station.StationSchedule == nil {
				station.StationSchedule = &[]bloodinfo.StationSchedule{schedule}
			} else {
				*station.StationSchedule = append(*station.StationSchedule, schedule)
			}
		}

		//Gorm save station, schedule, and status at this point. That's why we are taking schedule.Id only after.
		if err := tx.Save(&station).Error; err != nil {
			log.Printf("Error while saving station: %v", err)
			return err
		}

		if station.StationSchedule != nil {
			for _, schedule := range *station.StationSchedule {
				if schedule.Id != nil {
					schedulesIds = append(schedulesIds, *schedule.Id)
				}
			}
		}
	}

	var otherSchedules []bloodinfo.StationSchedule
	if err := tx.Not("id", schedulesIds).Find(&otherSchedules).Error; err != nil {
		log.Printf("Error while searching other schedules: %v", err)
		return err
	}
	for _, schedule := range otherSchedules {
		if isScheduleToday(schedule) {
			lastStatus := bloodinfo.StationStatus{}
			if err := tx.Where("station_schedule_id = ? AND user_id IS NULL", schedule.Id).Order("created_at DESC").First(&lastStatus).Error; err != nil {
				log.Printf("Error while searching status: %v", err)
				return err
			}
			if lastStatus.Id == nil || lastStatus.IsOpen {
				schedule.StationStatus = &[]bloodinfo.StationStatus{{IsOpen: false}}
				if err := tx.Save(&schedule).Error; err != nil {
					log.Printf("Error while saving status: %v", err)
					return err
				}
			}
		} else if !isDatePassed(schedule.Date) {
			if err := tx.Delete(&schedule).Error; err != nil {
				log.Printf("Error while deleting schedule: %v", err)
				return err
			}
		}
	}
	return nil
}

func isScheduleToday(schedule bloodinfo.StationSchedule) bool {
	currentDate := time.Now().UTC().Truncate(24 * time.Hour)
	scheduleDate := schedule.Date.UTC().Truncate(24 * time.Hour)
	return scheduleDate.Equal(currentDate)
}

func isDatePassed(date time.Time) bool {
	today := time.Now()
	return date.Year() < today.Year() || (date.Year() == today.Year() && date.YearDay() < today.YearDay())
}
