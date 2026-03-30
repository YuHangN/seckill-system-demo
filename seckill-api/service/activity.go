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
	db.AutoMigrate(&model.Activity{})
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
	s.rdb.Set(ctx, key, stock, 0)
	return a, nil
}

func (s *ActivityService) Get(ctx context.Context, id string) (*model.Activity, int64, error) {
	var a model.Activity
	if err := s.db.WithContext(ctx).First(&a, "id = ?", id).Error; err != nil {
		return nil, 0, err
	}
	// Read current stock from Redis.
	stock, _ := s.rdb.Get(ctx, fmt.Sprintf(stockKey, id)).Int64()
	return &a, stock, nil
}
