# [Tempsy API](https://tempsy.afifurrohman.my.id)

> Simple Temporary Files sharing RESTful API with oauth2
  
## API Documentation

  > :warning: **This API is not stable yet, maybe have some
  > breaking changes in the future.**

- [OpenAPI Specification](api/openapi-spec.yaml)

## Usage

### Requirements

- [x] WSL2 (Windows Subsystem for Linux)
  > Only need if you use Windows OS
- [x] Make (version >= 4.4.x)
- [x] Go (version >= 1.21.x)
- [x] Git (version >= 2.43.x)
- [ ] Docker (version >= 24.0.x)
  > Optional, only if you want to build image

### Installation

- Clone this repository

```sh
git clone https://github.com/afifurrohman-id/tempsy.git
```

- Go to project directory

```sh
cd tempsy
```

- Insert Variable into `.env` file

```sh
cat <<EOENV > configs/.env

# Server
GOOGLE_CLOUD_STORAGE_BUCKET=example-google-cloud-storage-bucket
APP_ENV=testing
PORT=3210
SERVER_URL=https://example.com

# Credentials
GOOGLE_CLOUD_STORAGE_SERVICE_ACCOUNT=JSON_GCP_SERVICE_ACCOUNT_CREDENTIAL
JWT_SECRET_KEY=example-jwt-secret-key

# Emulator
GOOGLE_CLOUD_STORAGE_EMULATOR_ENDPOINT=https://example.com/emulators/storage/v1

# testing
GOOGLE_OAUTH2_REFRESH_TOKEN_TEST=example-oauth2-refresh-token
GOOGLE_OAUTH2_CLIENT_ID_TEST=example-google-oauth2-client-id
GOOGLE_OAUTH2_CLIENT_SECRET_TEST=example-google-oauth2-client-secret

EOENV
```

- Download dependencies

```sh
go mod tidy
```

### Run

- Run Docker Compose

```sh
make compose-up
```

- Run Server

```sh
make run
```

- Build

```sh
make build
```

- Build Image

```sh
make build-image
```

- Test (Unit Test)

```sh
make test
```
