[![Build Status](https://github.com/il-blood-donation-info/blood-donation-backend/actions/workflows/on_push.yml/badge.svg)](https://github.com/il-blood-donation-info/blood-donation-backend/actions/workflows/on_push.yml)

# Blood Donation Backend Service
The project is used to collect and persist blood donation stations, and offers an API to manage stations availability.
It also allows querying for the stations.

View OpenAPI definition using Swagger UI [here](https://generator.swagger.io/?url=https://raw.githubusercontent.com/il-blood-donation-info/blood-donation-backend/main/pkg/api/openapi.yaml#/).

## Updating the API and generating code
In order to modify the API, edit the api/openapi.yaml file. Then, run the following commands to generate the code:
```bash
oapi-codegen -config pkg/api/api.cfg.yaml pkg/api/openapi.yaml
```

If `oapi-codegen` isn't available, install the code-gen dependency:
```bash
go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@latest
```

## Setting development environment

### Creating a Database
```postgresql
CREATE DATABASE bloodinfo;
CREATE USER mada WITH PASSWORD <your password>; # change this
GRANT ALL PRIVILEGES ON DATABASE bloodinfo TO mada;
```

### Running the server
```bash
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=mada
export DB_NAME=bloodinfo
export DB_PASSWORD= # your password
go run cmd/cert/tls-self-signed-cert.go
go run cmd/server/main.go
```

## Running using docker-compose

### Building the image and running the containers
Edit or change the db.env file to your liking. Then, run the following command:

```bash
docker-compose --env-file db.env up --build
```
### Stopping the containers
```bash
docker-compose --env-file db.env down --remove-orphans --volumes --rmi local
```

## Testing the API

### Creating a user
```bash
curl --cacert ./cert.pem -X POST https://localhost:8443/users -H "Content-Type: application/json" -d '{"description": "User description",
    "email": "user@example.com",
    "first_name": "John",
    "id": 1,
    "last_name": "Doe",
    "phone": "1234567890",
    "role": "Admin"}'
```

### Getting a user
```bash
curl --cacert ./cert.pem -s -X GET https://localhost:8443/users | jq
[
  {
    "description": "User description",
    "email": "user@example.com",
    "first_name": "John",
    "id": 1,
    "last_name": "Doe",
    "phone": "1234567890",
    "role": "Admin"
  }
]
```

### Running Scrapper test against real DB
```bash
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER_TEST=mada_test
export DB_NAME_TEST=bloodinfo_test
export DB_PASSWORD=mada

# Create the test database
TEST_DB_COMMANDS=$(cat <<EOF
CREATE DATABASE $DB_NAME_TEST;
CREATE USER $DB_USER_TEST WITH PASSWORD '$DB_PASSWORD';
GRANT ALL PRIVILEGES ON DATABASE $DB_NAME_TEST TO $DB_USER_TEST;
EOF
)

# Use echo to pass the commands to psql
echo "$TEST_DB_COMMANDS" | psql -U postgres

go test ./pkg/scraper/... -v
```
