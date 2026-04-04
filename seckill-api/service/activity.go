package service

import (
	"context"
	"fmt"
	"seckill-system/internal/model"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const stockKey = "stock:%s"

type ActivityService struct {
	db  *gorm.DB
	rdb *redis.Client
}

func NewActivityService(db *gorm.DB, rdb *redis.Client) *ActivityService {
	if err := db.AutoMigrate(&model.Activity{}); err != nil {
		panic("failed to migrate activity table: " + err.Error())
	}
	return &ActivityService{db: db, rdb: rdb}
}

func (s *ActivityService) Create(ctx context.Context, name string, stock int, start, end time.Time) (*model.Activity, error) {
	a := &model.Activity{
		ID:         uuid.New().String(),
		Name:       name,
		StockTotal: stock,
		StartTime:  start,
		EndTime:    end,
	}

	if err := s.db.WithContext(ctx).Create(a).Error; err != nil {
		return nil, err
	}

	key := fmt.Sprintf(stockKey, a.ID)
	if err := s.rdb.Set(ctx, key, stock, 0).Err(); err != nil {
		return nil, fmt.Errorf("activity created but failed to set redis stock: %w", err)
	}
	return a, nil
}

func (s *ActivityService) Get(ctx context.Context, id string) (*model.Activity, int64, error) {
	var a model.Activity
	if err := s.db.WithContext(ctx).First(&a, "id = ?", id).Error; err != nil {
		return nil, 0, err
	}
	// Read current stock from Redis.
	stock, err := s.rdb.Get(ctx, fmt.Sprintf(stockKey, id)).Int64()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get redis stock: %w", err)
	}
	return &a, stock, nil
}
