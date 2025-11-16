FROM golang:1.24-alpine3.21 AS builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download

WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o /bin/app ./cmd/server/main.go

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o /bin/migrate ./cmd/migrator/main.go

FROM scratch
COPY --from=builder /app/migrations /migrations
COPY --from=builder /bin/app /app
COPY --from=builder /bin/migrate /migrate
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

CMD ["/app"]
