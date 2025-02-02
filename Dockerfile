FROM golang:1.23-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o meido cmd/meido/main.go

FROM alpine:latest
WORKDIR /app

COPY --from=builder /app/meido .

COPY internal/database/migrations /app/database/migrations

RUN apk add --no-cache curl jq

RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.18.0/migrate.linux-amd64.tar.gz \
    | tar -xz -C /usr/local/bin
RUN chmod +x /usr/local/bin/migrate

COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
