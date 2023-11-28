FROM golang:alpine AS builder
LABEL authors="afif"
LABEL org.opencontainers.image.source=https://github.com/afifurrohman-id/tempsy-client
LABEL org.opencontainers.image.description="Tempsy Client Web App"
LABEL org.opencontainers.image.licenses=MIT

WORKDIR /src
COPY . .
RUN go mod tidy
RUN go fmt ./...

ENV CGO_ENABLED=0

# Reduce binary size by removing debug information
RUN go build -ldflags '-w -s' -o tempsy-client main.go

FROM alpine
WORKDIR /app

# Install CA certificates
RUN apk update && apk add --no-cache ca-certificates && update-ca-certificates

COPY --from=builder /src/tempsy-client .
COPY --from=builder /src/web ./web

CMD [ "./tempsy-client" ]
