package handler

import (
	"net/http"
	"seckill-system/internal/model"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type OrderHandler struct {
	db *gorm.DB
}

func NewOrderHandler(db *gorm.DB) *OrderHandler {
	return &OrderHandler{db: db}
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
	h.db.Order("created_at DESC").Limit(limit).Find(&orders)
	c.JSON(http.StatusOK, orders)
}
