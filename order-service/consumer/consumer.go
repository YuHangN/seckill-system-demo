package consumer

import (
	"context"
	"encoding/json"
	"log"
	"seckill-system/internal/kafka"
	"seckill-system/internal/model"
	"seckill-system/order-service/rabbitmq"
	"seckill-system/order-service/service"
)

type OrderConsumer struct {
	consumer    *kafka.Consumer
	orderSvc    *service.OrderService
	delayClient *rabbitmq.DelayClient
}

func New(c *kafka.Consumer, orderSvc *service.OrderService, delay *rabbitmq.DelayClient) *OrderConsumer {
	return &OrderConsumer{consumer: c, orderSvc: orderSvc, delayClient: delay}
}

func (oc *OrderConsumer) Run(ctx context.Context) {
	log.Println("order-service kafka consumer: started")

	for {
		msg, err := oc.consumer.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Printf("fetch error: %v", err)
			continue
		}

		if msg.Topic == "order.created" {
			var event model.OrderCreatedEvent
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				log.Printf("unmarshal error: %v", err)
			} else if err := oc.orderSvc.Create(ctx, event.OrderID, event.ActivityID, event.UserID); err != nil {
				log.Printf("create order error: %v", err)
			} else {
				// 发送延迟消息：30s 后若仍 PENDING 则取消
				oc.delayClient.PublishDelay(event.OrderID, event.ActivityID)
				log.Printf("order created: %s, timeout scheduled", event.OrderID)
			}
		}

		oc.consumer.CommitMessages(ctx, msg)
	}
}
