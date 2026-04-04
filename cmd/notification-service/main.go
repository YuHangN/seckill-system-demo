// cmd/stats-service/main.go
package main

import (
	"context"
	"seckill-system/internal/config"
	kafkaclient "seckill-system/internal/kafka"
	"seckill-system/notification-service/consumer"
	"strings"
)

func main() {
	cfg := config.Load()
	brokers := strings.Split(cfg.KafkaBrokers, ",")

	c := kafkaclient.NewConsumer(brokers, "stats-service-group",
		"order.created", "order.paid", "order.cancelled")
	defer c.Close()

	consumer.New(c).Run(context.Background())
}
