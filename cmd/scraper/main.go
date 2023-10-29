package main

import (
	"blood-donation-backend/pkg/scraper"
	"log"
	"net/http"
	"time"
)

func main() {
	db, err := scraper.InitDb()
	if err != nil {
		panic(err)
	}
	defer scraper.CloseDbConnection(db)

	s := scraper.Scraper{
		Client: &http.Client{},
	}
	madaResponse, err := s.ScrapeMada()
	if err != nil {
		log.Fatalf("Failed to scrape Mada: %s", err)
	}

	// Basic check if we got some data
	if len(madaResponse) == 0 {
		log.Fatal("Received empty response from Mada")
	}

	log.Println("SaveData")
	p := scraper.ScheduleDataWriter{
		DB:        db,
		SinceTime: time.Now(),
	}
	err = p.SaveData(madaResponse)
	if err != nil {
		log.Fatalf("Failed to SaveData: %s", err)
	}
}
