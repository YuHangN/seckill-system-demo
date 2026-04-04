# cmd/notification-service/Dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o notification-service ./cmd/notification-service

FROM alpine:3.19
COPY --from=builder /app/notification-service /notification-service
ENTRYPOINT ["/notification-service"]