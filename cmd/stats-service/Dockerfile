FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o stats-service ./cmd/stats-service

FROM alpine:3.19
COPY --from=builder /app/stats-service /stats-service
ENTRYPOINT ["/stats-service"]