// scraper/mada.go

package scraper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
	"os"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"blood-donation-backend/bloodinfo"
)

const customDateLayout = "2006-01-02"
const customDateTimeLayout = "2006-01-02 15:04"

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

func ConvertDonationToStation(d DonationDetail) bloodinfo.Station {
	openDateTime := d.DateDonation.Format(customDateLayout) + " " + d.FromHour  //timezone issue ?
	openTime, err := time.Parse(customDateTimeLayout, openDateTime)
	if err != nil {
		log.Fatal(err)
	}
	closeDateTime := d.DateDonation.Format(customDateLayout) + " " + d.ToHour
	closeTime, err := time.Parse(customDateTimeLayout, closeDateTime)
	if err != nil {
		log.Fatal(err)
	}

	return bloodinfo.Station{
		Address:   fmt.Sprintf("%s, %s %s", d.Street, d.NumHouse, d.City), //need formatting
		Name:      d.Name,   //need formatting (at least trim)
		OpenTime:  openTime,
		CloseTime: closeTime,
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
	client := &http.Client{}
	resp, err := client.Do(req)
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

func SaveData(donationDetails []DonationDetail)error{
    //todo : Clean DB before adding ? Filter for adding just stations for today ?

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
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal(err)
	}
	defer sqlDB.Close()

    if err != nil {
        return err
    }
	err = db.AutoMigrate(&bloodinfo.Station{})
	if err != nil {
		log.Fatal(err)
	}

    for _, donation := range donationDetails {
        station := ConvertDonationToStation(donation)
        result := db.Create(&station)
        if result.Error != nil {
            log.Fatal(result.Error)
        }
    }

	fmt.Println("Bulk insert completed successfully.")
	return nil
}