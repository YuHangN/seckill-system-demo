// Package db internal/db/client.go
package db

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func New(dsn string) *gorm.DB {
	// 获取数据库连接
	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		panic("postgres connection failed: " + err.Error())
	}
	return database
}
