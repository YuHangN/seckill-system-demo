package handler

import (
	"errors"
	"net/http"
	"seckill-system/seckill-api/service"

	"github.com/gin-gonic/gin"
)

type SeckillHandler struct {
	svc *service.SeckillService
}

func NewSeckillHandler(svc *service.SeckillService) *SeckillHandler {
	return &SeckillHandler{svc: svc}
}

type buyRequest struct {
	UserID     string `json:"user_id" binding:"required"`
	ActivityID string `json:"activity_id" binding:"required"`
}

func (h *SeckillHandler) Buy(c *gin.Context) {
	var req buyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	orderID, err := h.svc.Buy(c.Request.Context(), req.UserID, req.ActivityID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrSoldOut):
			c.JSON(http.StatusConflict, gin.H{"error": "sold_out"})
		case errors.Is(err, service.ErrDuplicate):
			c.JSON(http.StatusConflict, gin.H{"error": "duplicate_order"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	// 202 Accepted: request received, order is being processed asynchronously
	c.JSON(http.StatusAccepted, gin.H{"order_id": orderID})
}
