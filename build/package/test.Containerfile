FROM golang:1.20-alpine

WORKDIR /src

COPY . .

ENV CGO_ENABLED=0
ENV TZ=Asia/Jakarta

# Need install ca-certificates for tls compatibility for go library
RUN apk add --no-cache \
    ca-certificates=20230506-r1 && \
    update-ca-certificates

RUN go fix ./... && \
    go fmt ./... && \
    go vet ./... && \
    go mod tidy
