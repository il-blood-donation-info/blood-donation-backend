#!/bin/sh

# Run the scrapper in a loop with a 5-minute sleep
while true; do
  ./scraper
  sleep 300
done
