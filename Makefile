# Makefile
.PHONY: up down build logs dev-api

up:
	docker-compose up --build

down:
	docker-compose down -v

build:
	docker-compose build

logs:
	docker-compose logs -f

# Run seckill-api locally (for development)
dev-api:
	KAFKA_BROKERS=localhost:9092 REDIS_ADDR=localhost:6379 \
	DB_DSN="host=localhost user=postgres password=postgres dbname=seckill sslmode=disable" \
	PORT=8080 go run ./cmd/seckill-api