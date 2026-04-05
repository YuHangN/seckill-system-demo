// Package model internal/model/order.go
package model

import (
	"errors"
	"time"
)

var ErrOrderNotPending = errors.New("order is not pending")

type OrderStatus string

const (
	StatusPending   OrderStatus = "PENDING"
	StatusPaid      OrderStatus = "PAID"
	StatusCancelled OrderStatus = "CANCELLED"
)

type Order struct {
	ID         string      `json:"id" gorm:"primaryKey"`
	ActivityID string      `json:"activity_id" gorm:"index"`
	UserID     string      `json:"user_id"`
	Status     OrderStatus `json:"status"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
}
