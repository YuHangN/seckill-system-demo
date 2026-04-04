package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type StatsHandler struct {
	rdb *redis.Client
}

func NewStatsHandler(rdb *redis.Client) *StatsHandler {
	return &StatsHandler{rdb: rdb}
}

func (h *StatsHandler) GetStats(c *gin.Context) {
	ctx := c.Request.Context()

	keys := []string{
		"stats:total_requests",
		"stats:success_count",
		"stats:fail_count",
		"stats:cancel_count",
		"stats:paid_count",
	}
	vals, err := h.rdb.MGet(ctx, keys...).Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	toInt := func(v any) int64 {
		if v == nil {
			return 0
		}
		n, _ := strconv.ParseInt(v.(string), 10, 64)
		return n
	}

	c.JSON(http.StatusOK, gin.H{
		"total_requests": toInt(vals[0]),
		"success_count":  toInt(vals[1]),
		"fail_count":     toInt(vals[2]),
		"cancel_count":   toInt(vals[3]),
		"paid_count":     toInt(vals[4]),
	})
}

// GetQPS 获得最近 60 秒的 QPS 数据，返回一个数组，每个元素包含秒数和对应的请求数。
func (h *StatsHandler) GetQPS(c *gin.Context) {
	ctx := c.Request.Context()
	now := time.Now().Unix()

	type point struct {
		Second int64 `json:"second"`
		Count  int64 `json:"count"`
	}
	points := make([]point, 0, 60)

	for i := int64(59); i >= 0; i-- {
		sec := now - i
		key := fmt.Sprintf("stats:qps:%d", sec)
		val, _ := h.rdb.Get(ctx, key).Int64()
		points = append(points, point{Second: sec, Count: val})
	}

	c.JSON(http.StatusOK, points)
}
