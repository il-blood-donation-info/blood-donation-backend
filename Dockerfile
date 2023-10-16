# Use the official Go image as the base image
FROM golang:1.20-alpine as build

# Set the working directory
WORKDIR /app

# Copy the source code into the container
COPY . .

# Build the Go application
RUN go build -o blood-info .

FROM alpine
WORKDIR /app

COPY --from=build /app/./blood-info blood-info

RUN addgroup -S gouser && adduser -S gouser -G gouser
USER gouser

# Expose the port that the application will run on
EXPOSE 8080

# Command to run the Go application
CMD ["./blood-info"]
