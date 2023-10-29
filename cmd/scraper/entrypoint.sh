#!/bin/sh

# Check if interval provided or set default to 5 minutes
if [ -z "$SCRAPER_INTERVAL" ]; then
  SCRAPER_INTERVAL=300
fi

# Run the scrapper in a loop with a 5-minute sleep
while true; do
  ./scraper
  sleep $SCRAPER_INTERVAL
done
