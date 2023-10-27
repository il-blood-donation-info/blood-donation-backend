# Use the official Go image as the base image
FROM golang:1.20-alpine as build

# Set the working directory
WORKDIR /app

# Copy the source code into the container
COPY . .

# Build the Go application
RUN go build -o blood-info .

# Generate self-signed certificate
RUN go build -o tls-self-signed-cert cert/tls-self-signed-cert.go

FROM alpine
WORKDIR /app

COPY --from=build /app/./blood-info blood-info
COPY --from=build /app/./tls-self-signed-cert tls-self-signed-cert

# FIXME: This is a hack to get the self-signed certificate to work. There must be a better way.
RUN ./tls-self-signed-cert

RUN addgroup -S gouser && adduser -S gouser -G gouser
RUN chown -R gouser:gouser ./cert.pem ./key.pem
USER gouser

# Expose the port that the application will run on
EXPOSE 8080

# Command to run the Go application
CMD ["./blood-info", "--certfile", "./cert.pem", "--keyfile", "./key.pem"]
