# Build stage
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/server/

# Run stage
FROM alpine:3.19
WORKDIR /app
RUN apk add --no-cache ca-certificates curl
COPY --from=builder /app/main .
COPY app.env .
COPY internal/db/migration ./internal/db/migration
COPY docs/swagger ./docs/swagger
EXPOSE 8080 9090
CMD ["/app/main"]
