package service

import (
	"context"
	"seckill-system/internal/model"

	"gorm.io/gorm"
)

type OrderService struct {
	db *gorm.DB
}

func NewOrderService(db *gorm.DB) *OrderService {
	db.AutoMigrate(&model.Order{})
	return &OrderService{db: db}
}

// Create 创建订单
func (s *OrderService) Create(ctx context.Context, id, activityID, userID string) error {
	return s.db.WithContext(ctx).Create(&model.Order{
		ID:         id,
		ActivityID: activityID,
		UserID:     userID,
		Status:     model.StatusPending,
	}).Error
}

func (s *OrderService) UpdateStatus(ctx context.Context, id string, status model.OrderStatus) error {
	return s.db.WithContext(ctx).
		Model(&model.Order{}).
		Where("id = ?", id).
		Update("status", status).Error
}

func (s *OrderService) GetStatus(ctx context.Context, id string) (model.OrderStatus, error) {
	var o model.Order
	err := s.db.WithContext(ctx).Select("status").First(&o, "id = ?", id).Error
	return o.Status, err
}
