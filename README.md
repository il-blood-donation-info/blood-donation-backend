# Updating the API and generating code
```bash
go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@latest
oapi-codegen -config server.cfg.yaml openapi.yaml
```

# Create a Database
```postgresql
CREATE DATABASE bloodinfo;
CREATE USER mada WITH PASSWORD 'mada';
GRANT ALL PRIVILEGES ON DATABASE bloodinfo TO mada;
```

