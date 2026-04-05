package handler

import (
	"log"
	"net/http"
	"seckill-system/internal/kafka"
	"seckill-system/internal/model"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type OrderHandler struct {
	db       *gorm.DB
	producer *kafka.Producer
}

func NewOrderHandler(db *gorm.DB, producer *kafka.Producer) *OrderHandler {
	return &OrderHandler{db: db, producer: producer}
}

func (h *OrderHandler) GetOrder(c *gin.Context) {
	id := c.Param("id")
	var order model.Order
	if err := h.db.First(&order, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}
	c.JSON(http.StatusOK, order)
}

func (h *OrderHandler) GetRecentOrders(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	var orders []model.Order
	if err := h.db.Order("created_at DESC").Limit(limit).Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, orders)
}

func (h *OrderHandler) Pay(c *gin.Context) {
	id := c.Param("id")
	ctx := c.Request.Context()

	var order model.Order
	if err := h.db.WithContext(ctx).First(&order, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}

	// 只有 PENDING 状态才能支付
	if order.Status != model.StatusPending {
		c.JSON(http.StatusBadRequest, gin.H{"error": "order is not pending"})
		return
	}

	// 更新状态为 PAID
	if err := h.db.WithContext(ctx).Model(&order).Update("status", model.StatusPaid).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 发布 Kafka 事件，通知 stats-service 和 notification-service
	event := model.OrderPaidEvent{OrderID: id}
	if err := h.producer.Publish(ctx, "order.paid", event); err != nil {
		// 支付已成功，Kafka 失败只记日志，不回滚
		log.Printf("pay: failed to publish order.paid event: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{"order_id": id, "status": "paid"})
}
