FROM golang:alpine AS builder
LABEL authors=afif

WORKDIR /src
COPY . .

# Install CA certificates and update them
RUN apk update && apk add --no-cache ca-certificates && update-ca-certificates

RUN go mod tidy && go fmt ./...
ENV CGO_ENABLED=0

# Reduce binary size by removing debug information
RUN go build -ldflags '-w -s' -o tempsy main.go

FROM scratch
WORKDIR /app

COPY --from=builder /src/tempsy .
# Copy CA certificates
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

CMD [ "./tempsy" ]