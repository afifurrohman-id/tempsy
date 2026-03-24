FROM golang:alpine AS builder
WORKDIR /src

RUN apk add --no-cache ca-certificates && \
    update-ca-certificates

ENV CGO_ENABLED=0

COPY go.* .
RUN go mod download

COPY . .

RUN go build \
    -ldflags "-w -s" \
    -o grpc \
    cmd/server/main.go

FROM scratch
LABEL org.opencontainers.image.authors="afif"
LABEL org.opencontainers.image.licenses="MIT"
WORKDIR /app

COPY --from=builder /src/grpc .
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

CMD [ "./grpc" ]
