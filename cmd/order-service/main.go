package main

import (
	"context"
	"seckill-system/internal/config"
	"seckill-system/internal/db"
	"seckill-system/internal/kafka"
	"seckill-system/internal/redis"
	"seckill-system/order-service/consumer"
	rmq "seckill-system/order-service/rabbitmq"
	"seckill-system/order-service/service"
	"strings"
)

func main() {
	cfg := config.Load()
	brokers := strings.Split(cfg.KafkaBrokers, ",")
	ctx := context.Background()

	rdb := redis.New(cfg.RedisAddr)
	database := db.New(cfg.DBDSN)
	producer := kafka.NewProducer(brokers)
	defer producer.Close()

	orderSvc := service.NewOrderService(database)

	// RabbitMQ DLX client
	delayClient, err := rmq.New(cfg.RabbitMQUrl, orderSvc, rdb, producer)
	if err != nil {
		panic("rabbitmq init failed: " + err.Error())
	}
	defer delayClient.Close()

	// RabbitMQ consumer 在后台运行：监听 order.cancel 死信队列
	go delayClient.ConsumeCancel(ctx)

	// Kafka consumer 在前台运行：处理 order.created 事件
	c := kafka.NewConsumer(brokers, "order-service-group", "order.created")
	defer c.Close()

	consumer.New(c, orderSvc, delayClient).Run(ctx)
}
