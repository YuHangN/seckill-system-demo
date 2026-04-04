package consumer

import (
	"context"
	"encoding/json"
	"log"
	"seckill-system/internal/kafka"
	"seckill-system/internal/model"
)

type NotificationConsumer struct {
	consumer *kafka.Consumer
}

func New(c *kafka.Consumer) *NotificationConsumer {
	return &NotificationConsumer{consumer: c}
}

func (nc *NotificationConsumer) Run(ctx context.Context) {
	log.Println("notification-service: starting consumer")

	for {
		msg, err := nc.consumer.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			continue
		}

		switch msg.Topic {
		case "order.created":
			var e model.OrderCreatedEvent
			if err := json.Unmarshal(msg.Value, &e); err != nil {
				log.Printf("[NOTIFY] unmarshal order.created error: %v", err)
			} else {
				log.Printf("[NOTIFY] Order created - user=%s order=%s", e.UserID, e.OrderID)
			}
		case "order.paid":
			var e model.OrderPaidEvent
			if err := json.Unmarshal(msg.Value, &e); err != nil {
				log.Printf("[NOTIFY] unmarshal order.paid error: %v", err)
			} else {
				log.Printf("[NOTIFY] Payment confirmed - order=%s", e.OrderID)
			}
		case "order.cancelled":
			var e model.OrderCancelledEvent
			if err := json.Unmarshal(msg.Value, &e); err != nil {
				log.Printf("[NOTIFY] unmarshal order.cancelled error: %v", err)
			} else {
				log.Printf("[NOTIFY] Order cancelled - order=%s", e.OrderID)
			}
		}

		nc.consumer.CommitMessages(ctx, msg)
	}
}
