// cmd/notification-service/main.go
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

	if err := kafkaclient.EnsureTopics(brokers, []string{"order.created", "order.paid", "order.cancelled"}); err != nil {
		panic("ensure kafka topics: " + err.Error())
	}

	c := kafkaclient.NewConsumer(brokers, "notification-service-group",
		"order.created", "order.paid", "order.cancelled")
	defer c.Close()

	consumer.New(c).Run(context.Background())
}
