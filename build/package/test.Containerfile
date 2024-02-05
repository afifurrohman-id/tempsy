FROM golang:1.20-alpine

WORKDIR /src

COPY . .

ENV CGO_ENABLED=0
ENV TZ=Asia/Jakarta

# Need install ca-certificates for tls compatibility for go library
# hadolint ignore=DL3018
RUN apk add --no-cache \
    ca-certificates && \
    update-ca-certificates

RUN go fix ./... && \
    go fmt ./... && \
    go vet ./... && \
    go mod tidy
