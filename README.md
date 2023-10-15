# Updating the API and generating code
```bash
go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@latest
oapi-codegen -config server.cfg.yaml openapi.yaml
```

# Create a Database
```postgresql
CREATE DATABASE bloodinfo;
CREATE USER mada WITH PASSWORD <your password>; # change this
GRANT ALL PRIVILEGES ON DATABASE bloodinfo TO mada;
```

# Run the server
```bash
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=mada
export DB_NAME=bloodinfo
export DB_PASSWORD= # your password
go run main.go
```
