package service

import (
	"context"
	"errors"
	"fmt"
	"log"
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

	// 2. Dedup check: SET NX with 5-minute TTL TODO：添加限制，每个用户只能购买一次同一活动的商品
	dedupKey := fmt.Sprintf(dedupKeyFmt, userID, activityID)
	// SET key value NX PX 300000 — 单条原子命令，返回 "OK" 表示设置成功
	res, err := s.rdb.SetArgs(ctx, dedupKey, "1", redis.SetArgs{
		Mode: "NX",
		TTL:  5 * time.Minute,
	}).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return "", err // 真正的 Redis 错误
	}
	// key 已存在时 NX 不设置，res="" 且 err=redis.Nil
	set := res == "OK"
	// 如果没有设置成功，说明用户点击太快或者重复提交了订单，直接返回重复错误。
	if !set {
		s.rdb.Incr(ctx, "stats:fail_count")
		return "", ErrDuplicate
	}

	// 3. Atomic stock deduction
	stockKey := fmt.Sprintf("stock:%s", activityID)
	remaining, err := s.rdb.Decr(ctx, stockKey).Result()
	if err != nil {
		if delErr := s.rdb.Del(ctx, dedupKey).Err(); delErr != nil {
			log.Printf("rollback: failed to delete dedup key: %v", delErr)
		}
		return "", err
	}
	// 如果扣库存后剩余量小于 0，说明库存不足了，订单失败。
	if remaining < 0 {
		// Undo the decrement — item is sold out
		if err := s.rdb.Incr(ctx, stockKey).Err(); err != nil {
			log.Printf("rollback: failed to restore stock: %v", err)
		}
		if err := s.rdb.Del(ctx, dedupKey).Err(); err != nil {
			log.Printf("rollback: failed to delete dedup key: %v", err)
		}
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
		if rbErr := s.rdb.Incr(ctx, stockKey).Err(); rbErr != nil {
			log.Printf("rollback: failed to restore stock: %v", rbErr)
		}
		if rbErr := s.rdb.Del(ctx, dedupKey).Err(); rbErr != nil {
			log.Printf("rollback: failed to delete dedup key: %v", rbErr)
		}
		return "", err
	}

	s.rdb.Incr(ctx, "stats:success_count")
	return orderID, nil
}
