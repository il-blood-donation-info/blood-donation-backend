package scraper

import (
	"fmt"
	"testing"
)

func TestScrapeMada(t *testing.T) {
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
}
