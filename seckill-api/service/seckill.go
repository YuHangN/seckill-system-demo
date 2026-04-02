package service

import (
	"context"
	"errors"
	"fmt"
	"seckill-system/internal/model"
	"time"

	"seckill-system/internal/kafka"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

var (
	ErrSoldOut   = errors.New("sold_out")
	ErrDuplicate = errors.New("duplicate_order")
)

const (
	TopicOrderCreated   = "order.created"
	TopicOrderPaid      = "order.paid"
	TopicOrderCancelled = "order.cancelled"

	dedupKeyFmt = "order:lock:%s:%s" // order:lock:{user_id}:{activity_id}
	qpsKeyFmt   = "stats:qps:%d"     // stats:qps:{unix_second}
)

type SeckillService struct {
	rdb      *redis.Client
	producer *kafka.Producer
}

func NewSeckillService(rdb *redis.Client, producer *kafka.Producer) *SeckillService {
	return &SeckillService{rdb: rdb, producer: producer}
}

func (s *SeckillService) Buy(ctx context.Context, userID, activityID string) (string, error) {
	// 1. Track QPS: INCR stats:qps:{current_second}, expire after 120s
	qpsKey := fmt.Sprintf(qpsKeyFmt, time.Now().Unix())
	s.rdb.Incr(ctx, qpsKey)
	s.rdb.Expire(ctx, qpsKey, 120*time.Second)
	s.rdb.Incr(ctx, "stats:total_requests")

	// 2. Dedup check: SET NX with 5-minute TTL
	dedupKey := fmt.Sprintf(dedupKeyFmt, userID, activityID)
	// SET key value NX PX 300000 — 单条原子命令，返回 "OK" 表示设置成功
	res, err := s.rdb.SetArgs(ctx, dedupKey, "1", redis.SetArgs{
		Mode: "NX",
		TTL:  5 * time.Minute,
	}).Result()
	set := res == "OK"

	if err != nil {
		return "", err
	}
	// 如果没有设置成功，说明用户点击太快或者重复提交了订单，直接返回重复错误。
	if !set {
		s.rdb.Incr(ctx, "stats:fail_count")
		return "", ErrDuplicate
	}

	// 3. Atomic stock deduction
	stockKey := fmt.Sprintf("stock:%s", activityID)
	remaining, err := s.rdb.Decr(ctx, stockKey).Result()
	if err != nil {
		s.rdb.Del(ctx, dedupKey)
		return "", err
	}
	// 如果扣库存后剩余量小于 0，说明库存不足了，订单失败。
	if remaining < 0 {
		// Undo the decrement — item is sold out
		s.rdb.Incr(ctx, stockKey)
		s.rdb.Del(ctx, dedupKey)
		s.rdb.Incr(ctx, "stats:fail_count")
		return "", ErrSoldOut
	}

	// 4. Produce Kafka event
	orderID := uuid.New().String()
	event := model.OrderCreatedEvent{
		OrderID:    orderID,
		ActivityID: activityID,
		UserID:     userID,
	}
	if err := s.producer.Publish(ctx, "order.created", event); err != nil {
		s.rdb.Incr(ctx, stockKey)
		s.rdb.Del(ctx, dedupKey)
		return "", err
	}

	s.rdb.Incr(ctx, "stats:success_count")
	return orderID, nil
}
