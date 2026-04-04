package main

import (
	"seckill-system/internal/config"
	"seckill-system/internal/db"
	"seckill-system/internal/kafka"
	"seckill-system/internal/redis"
	"seckill-system/seckill-api/handler"
	"seckill-system/seckill-api/middleware"
	"seckill-system/seckill-api/service"
	"strings"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	rdb := redis.New(cfg.RedisAddr)
	database := db.New(cfg.DBDSN)
	brokers := strings.Split(cfg.KafkaBrokers, ",")

	if err := kafka.EnsureTopics(brokers, []string{"order.created", "order.paid", "order.cancelled"}); err != nil {
		panic("ensure kafka topics: " + err.Error())
	}

	producer := kafka.NewProducer(brokers)
	defer producer.Close()

	// Services
	activitySvc := service.NewActivityService(database, rdb)
	seckillSvc := service.NewSeckillService(rdb, producer)

	// Handlers
	activityH := handler.NewActivityHandler(activitySvc)
	seckillH := handler.NewSeckillHandler(seckillSvc)
	statsH := handler.NewStatsHandler(rdb)
	orderH := handler.NewOrderHandler(database)

	r := gin.Default()
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	api := r.Group("/api")
	{
		api.POST("/activity", activityH.Create)
		api.GET("/activity/:id", activityH.Get)

		// Rate limit: 10 req/s per IP on the seckill endpoint
		api.POST("/seckill", middleware.NewRateLimiter(10, 10).Middleware(), seckillH.Buy)
		api.GET("/order/:id", orderH.GetOrder)
		api.GET("/orders/recent", orderH.GetRecentOrders)

		api.GET("/stats", statsH.GetStats)
		api.GET("/stats/qps", statsH.GetQPS)
	}

	r.Run(":" + cfg.Port)
}
