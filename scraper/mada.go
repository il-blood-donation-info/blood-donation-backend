// scraper/mada.go

package scraper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const customDateLayout = "2006-01-02"

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
