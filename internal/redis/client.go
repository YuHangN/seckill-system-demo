// Package redisclient internal/redis/client.go
package redisclient

import (
	"context"

	"github.com/redis/go-redis/v9"
)

func New(addr string) *redis.Client {
	rdb := redis.NewClient(&redis.Options{Addr: addr})
	// Verify connection
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		panic("redis connection failed: " + err.Error())
	}
	return rdb
}
