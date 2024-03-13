FROM golang:1.20-alpine

WORKDIR /src

# Need install ca-certificates for tls compatibility for go library
# hadolint ignore=DL3018
RUN apk add --no-cache \
    ca-certificates && \
    update-ca-certificates

ENV CGO_ENABLED=0
ENV TZ=Asia/Jakarta

# Cache layer for go dependencies
COPY go.* .
RUN go mod download

COPY . .
RUN go fix ./... && \
    go fmt ./... && \
    go vet ./... && \
    go mod tidy
