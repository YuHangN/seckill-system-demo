// cmd/stats-service/main.go
package main

import (
	"context"
	"seckill-system/internal/config"
	kafkaclient "seckill-system/internal/kafka"
	redisclient "seckill-system/internal/redis"
	"seckill-system/stats-service/consumer"
	"strings"
)

func main() {
	cfg := config.Load()
	brokers := strings.Split(cfg.KafkaBrokers, ",")

	rdb := redisclient.New(cfg.RedisAddr)
	c := kafkaclient.NewConsumer(brokers, "stats-service-group",
		"order.created", "order.paid", "order.cancelled")
	defer c.Close()

	consumer.New(c, rdb).Run(context.Background())
}
