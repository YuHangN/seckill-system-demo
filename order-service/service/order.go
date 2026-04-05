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
	if err := db.AutoMigrate(&model.Order{}); err != nil {
		panic("failed to migrate order table: " + err.Error())
	}
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
	// 只允许从 PENDING 状态变更，防止并发时取消覆盖已支付的订单
	result := s.db.WithContext(ctx).
		Model(&model.Order{}).
		Where("id = ? AND status = ?", id, model.StatusPending).
		Update("status", status)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return model.ErrOrderNotPending
	}
	return nil
}

func (s *OrderService) GetStatus(ctx context.Context, id string) (model.OrderStatus, error) {
	var o model.Order
	err := s.db.WithContext(ctx).Select("status").First(&o, "id = ?", id).Error
	return o.Status, err
}
