package handler

import (
	"net/http"
	"seckill-system/seckill-api/service"
	"time"

	"github.com/gin-gonic/gin"
)

type ActivityHandler struct {
	svc *service.ActivityService
}

func NewActivityHandler(svc *service.ActivityService) *ActivityHandler {
	return &ActivityHandler{svc: svc}
}

type createActivityRequest struct {
	Name       string    `json:"name" binding:"required"`
	StockTotal int       `json:"stock_total" binding:"required,min=1"`
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`
}

func (h *ActivityHandler) Create(c *gin.Context) {
	var req createActivityRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	a, err := h.svc.Create(c.Request.Context(), req.Name, req.StockTotal, req.StartTime, req.EndTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, a)
}

// Get GetActivity 获取活动详情和当前库存
func (h *ActivityHandler) Get(c *gin.Context) {
	id := c.Param("id")

	a, stock, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "activity not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"activity": a, "current_stock": stock})
}
