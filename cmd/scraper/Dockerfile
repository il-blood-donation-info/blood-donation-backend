# Use the official Go image as the base image
FROM golang:1.20-alpine as build

# Set the working directory
WORKDIR /app

# Copy the source code into the container
COPY . .

# Build the Go application
RUN go build -o scraper /app/cmd/scraper

FROM alpine
WORKDIR /app

COPY --from=build /app/scraper ./scraper
COPY --from=build /app/cmd/scraper/entrypoint.sh ./entrypoint.sh
RUN chmod +x ./entrypoint.sh

RUN addgroup -S gouser && adduser -S gouser -G gouser
USER gouser

# Command to run the scraper every 5 minutes
CMD ["./entrypoint.sh", "$SCRAPER_INTERVAL"]

