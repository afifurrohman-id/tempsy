FROM golang:alpine AS builder
LABEL authors=afif

WORKDIR /src
COPY . .

RUN go mod tidy && go fmt ./...
ENV CGO_ENABLED=0
# Reduce binary size by removing debug information
RUN go build -ldflags '-w -s' -o tempsy main.go

FROM alpine
WORKDIR /app

# Install CA certificates
RUN apk update && apk add --no-cache ca-certificates && update-ca-certificates

COPY --from=builder /src/tempsy .
CMD [ "./tempsy" ]