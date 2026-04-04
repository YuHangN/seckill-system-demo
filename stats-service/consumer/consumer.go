package consumer

import (
	"context"
	"encoding/json"
	"log"
	"seckill-system/internal/kafka"
	"seckill-system/internal/model"

	"github.com/redis/go-redis/v9"
)

type StatsConsumer struct {
	consumer *kafka.Consumer
	rdb      *redis.Client
}

func New(c *kafka.Consumer, rdb *redis.Client) *StatsConsumer {
	return &StatsConsumer{consumer: c, rdb: rdb}
}

func (sc *StatsConsumer) Run(ctx context.Context) {
	log.Println("stats-service: starting consumer")

	for {
		msg, err := sc.consumer.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			continue
		}

		switch msg.Topic {
		case "order.cancelled":
			var event model.OrderCancelledEvent
			if err := json.Unmarshal(msg.Value, &event); err == nil {
				sc.rdb.Incr(ctx, "stats:cancel_count")
			}
		case "order.paid":
			var event model.OrderPaidEvent
			if err := json.Unmarshal(msg.Value, &event); err == nil {
				sc.rdb.Incr(ctx, "stats:paid_count")
			}
		}

		sc.consumer.CommitMessages(ctx, msg)
	}
}
