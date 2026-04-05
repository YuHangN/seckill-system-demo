package rabbitmq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"seckill-system/internal/kafka"
	"seckill-system/internal/model"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
)

const (
	dlxExchange   = "order.dlx"
	delayQueue    = "order.delay"
	cancelQueue   = "order.cancel"
	cancelRouting = "order.cancel"
	ttlMs         = 30000 // 30 秒
)

// DelayClient 封装 RabbitMQ 连接，负责设置拓扑、发送延迟消息、消费死信队列
type DelayClient struct {
	conn     *amqp.Connection
	ch       *amqp.Channel
	orderSvc OrderStatusChecker
	rdb      *redis.Client
	producer *kafka.Producer
}

// OrderStatusChecker 是 order-service 提供的接口，避免循环引用
type OrderStatusChecker interface {
	GetStatus(ctx context.Context, id string) (model.OrderStatus, error)
	UpdateStatus(ctx context.Context, id string, status model.OrderStatus) error
}

func New(url string, orderSvc OrderStatusChecker, rdb *redis.Client, producer *kafka.Producer) (*DelayClient, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("rabbitmq dial: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("rabbitmq channel: %w", err)
	}

	if err := setupTopology(ch); err != nil {
		return nil, err
	}

	return &DelayClient{conn: conn, ch: ch, orderSvc: orderSvc, rdb: rdb, producer: producer}, nil
}

// setupTopology 创建交换机、队列和绑定关系
func setupTopology(ch *amqp.Channel) error {
	// 1. 创建死信交换机，用于接收 order.delay 的过期消息
	if err := ch.ExchangeDeclare(dlxExchange, "direct", true, false, true, false, nil); err != nil {
		return fmt.Errorf("declare dlx exchange: %w", err)
	}

	// 2. Delay queue：消息在这里等待 TTL 过期后转发到 DLX 并附带 routing key
	if _, err := ch.QueueDeclare(delayQueue, true, false, false, false, amqp.Table{
		"x-message-ttl":             int32(ttlMs),
		"x-dead-letter-exchange":    dlxExchange,
		"x-dead-letter-routing-key": cancelRouting,
	}); err != nil {
		return fmt.Errorf("declare delay queue: %w", err)
	}

	// 3. Cancel queue：接收从 DLX 路由过来的死信
	if _, err := ch.QueueDeclare(cancelQueue, true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare cancel queue: %w", err)
	}

	// 4. 将 cancel queue 绑定到 DLX
	return ch.QueueBind(cancelQueue, cancelRouting, dlxExchange, false, nil)
}

// PublishDelay 向 order.delay 队列发送一条消息，30s 后自动路由到 order.cancel
func (c *DelayClient) PublishDelay(orderID, activityID string) error {
	body, err := json.Marshal(model.OrderCancelledEvent{
		OrderID:    orderID,
		ActivityID: activityID,
	})
	if err != nil {
		return fmt.Errorf("marshal delay message: %w", err)
	}

	return c.ch.Publish("", delayQueue, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent, // 消息持久化，broker 重启不丢失
		Body:         body,
	})
}

func (c *DelayClient) ConsumeCancel(ctx context.Context) {
	msgs, err := c.ch.Consume(cancelQueue, "", false, false, false, false, nil)
	if err != nil {
		log.Fatalf("rabbitmq consume: %v", err)
	}
	log.Println("order-service rabbitmq: waiting for timeout messages")

	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-msgs:
			if !ok {
				return
			}
			c.handleCancel(ctx, msg)
		}
	}
}

func (c *DelayClient) handleCancel(ctx context.Context, msg amqp.Delivery) {
	var event model.OrderCancelledEvent
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		log.Printf("rabbitmq: unmarshal error: %v", err)
		msg.Nack(false, false) // 丢弃无法解析的消息
		return
	}

	// 条件更新：只有 PENDING 状态才能取消，防止并发时覆盖已支付的订单
	if err := c.orderSvc.UpdateStatus(ctx, event.OrderID, model.StatusCancelled); err != nil {
		if errors.Is(err, model.ErrOrderNotPending) {
			// 订单已支付或已取消，正常跳过
			log.Printf("rabbitmq: order %s already handled, skip cancel", event.OrderID)
			msg.Ack(false)
			return
		}
		log.Printf("rabbitmq: cancel order error: %v", err)
		msg.Nack(false, true) // requeue
		return
	}

	c.rdb.Incr(ctx, fmt.Sprintf("stock:%s", event.ActivityID))
	// 通知 Kafka：stats/notification 服务需要感知这个取消事件
	if err := c.producer.Publish(ctx, "order.cancelled", event); err != nil {
		log.Printf("rabbitmq: failed to publish order.cancelled event: %v", err)
	}

	msg.Ack(false)
	log.Printf("rabbitmq: order %s cancelled via timeout", event.OrderID)
}

func (c *DelayClient) Close() {
	c.ch.Close()
	c.conn.Close()
}
