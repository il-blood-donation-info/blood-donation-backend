# Updating the API and generating code
In order to modify the API, edit the openapi.yaml file. Then, run the following commands to generate the code:
```bash
oapi-codegen -config server.cfg.yaml openapi.yaml
```

## Installing the code-gen dependency
```bash
go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@latest
```

# Creating a Database
```postgresql
CREATE DATABASE bloodinfo;
CREATE USER mada WITH PASSWORD <your password>; # change this
GRANT ALL PRIVILEGES ON DATABASE bloodinfo TO mada;
```

# Running the server
```bash
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=mada
export DB_NAME=bloodinfo
export DB_PASSWORD= # your password
go run main.go
```
