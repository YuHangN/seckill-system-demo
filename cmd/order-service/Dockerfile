# cmd/order-service/Dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o order-service ./cmd/order-service

FROM alpine:3.19
COPY --from=builder /app/order-service /order-service
ENTRYPOINT ["/order-service"]