# cmd/notification-service/Dockerfile
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /bin/app ./cmd/notification-service

FROM alpine:3.19
COPY --from=builder /bin/app /notification-service
ENTRYPOINT ["/notification-service"]