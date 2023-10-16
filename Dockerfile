# Use the official Go image as the base image
FROM golang:1.20-alpine

# Set the working directory
WORKDIR /app

# Copy the source code into the container
COPY . .

# Build the Go application
RUN go build -o blood-info .

# Expose the port that the application will run on
EXPOSE 8080

# Command to run the Go application
CMD ["./blood-info"]