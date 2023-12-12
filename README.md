# [Tempsy API](https://api.tempsy.afifurrohman.my.id)
> Simple Temporary Files sharing RESTful API with oauth2
  
## API Documentation
> :warning: **This API is not stable yet, maybe have some breaking changes in the future.**
- [OpenAPI Specification](api/openapi-spec.yaml)

## Usage

### Requirements

- [x] Git Bash for Windows (version >= 2.41.x)
  > only need if you're using windows OS
- [x] Go (version >= 1.21.x)
- [x] Docker (version >= 24.0.x)

### Installation
- Clone this repository

```sh
git clone https://github.com/afifurrohman-id/tempsy.git
```

- Go to project directory

```sh
cd tempsy
```

- Create `.env` file

```sh
touch configs/.env
```

- Insert Variable into `.env` file

```sh
# Server
GOOGLE_CLOUD_STORAGE_BUCKET=example-google-cloud-storage-bucket
APP_ENV=testing
PORT=3210
SERVER_URI=https://example.com

# Credentials
GOOGLE_CLOUD_STORAGE_SERVICE_ACCOUNT=JSON_GCP_SERVICE_ACCOUNT_CREDENTIAL
JWT_SECRET_KEY=example-jwt-secret-key

# Emulator
GOOGLE_CLOUD_STORAGE_EMULATOR_ENDPOINT=https://example.com/emulators/storage/v1

# testing
GOOGLE_OAUTH2_REFRESH_TOKEN_TEST=example-oauth2-refresh-token
GOOGLE_OAUTH2_CLIENT_ID_TEST=example-google-oauth2-client-id
GOOGLE_OAUTH2_CLIENT_SECRET_TEST=example-google-oauth2-client-secret
```

- Download dependencies

```sh
go mod tidy
```

### Run
- Run Docker Compose

```sh
docker compose -f deployments/compose.yaml up -d
```

- Run Server

```sh
go run main.go
```

- Build
```sh
go build -o tempsy main.go
```

- Build Image

```sh
docker build -f build/package/Containerfile -t tempsy .
```

- Test (Unit Test)

```sh
go test -v --cover ./...
```
