FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY ../../seckill-api .
RUN CGO_ENABLED=0 go build -o seckill-api ./cmd/seckill-api

FROM alpine:3.19
COPY --from=builder /app/seckill-api /seckill-api
EXPOSE 8080
ENTRYPOINT ["/seckill-api"]