// Package model internal/model/activity.go
package model

import "time"

type Activity struct {
	ID         string    `json:"id" gorm:"primaryKey"`
	Name       string    `json:"name"`
	StockTotal int       `json:"stock_total"`
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`
	CreatedAt  time.Time `json:"created_at"`
}
