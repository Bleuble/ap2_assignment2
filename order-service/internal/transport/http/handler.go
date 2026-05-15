package http

import (
	"fmt"
	"net/http"
	"order-service/internal/usecase"

	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	useCase *usecase.OrderUseCase
}

func NewOrderHandler(uc *usecase.OrderUseCase) *OrderHandler {
	return &OrderHandler{useCase: uc}
}

func (h *OrderHandler) RegisterRoutes(router *gin.Engine) {
	router.POST("/orders", h.CreateOrder)
	router.GET("/orders/:id", h.GetOrder)
	router.PATCH("/orders/:id/cancel", h.CancelOrder)
	router.GET("/orders/filter", h.GetOrdersByRange)
	router.GET("/payments", h.ListPayments)
}

type CreateOrderRequest struct {
	CustomerID string `json:"customer_id"`
	ItemName   string `json:"item_name"`
	Amount     int64  `json:"amount"`
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid format"})
		return
	}

	idempotencyKey := c.GetHeader("Idempotency-Key")

	order, err := h.useCase.CreateOrder(req.CustomerID, req.ItemName, req.Amount, idempotencyKey)
	if err != nil {
		if err.Error() == "customer_id is required" || err.Error() == "item_name is required" || err.Error() == "amount must be greater than zero" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, order)
}

func (h *OrderHandler) GetOrder(c *gin.Context) {
	id := c.Param("id")
	order, err := h.useCase.GetOrder(id)
	if err != nil || order == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}

	c.JSON(http.StatusOK, order)
}

func (h *OrderHandler) CancelOrder(c *gin.Context) {
	id := c.Param("id")
	if err := h.useCase.CancelOrder(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "order cancelled successfully"})
}

func (h *OrderHandler) GetOrdersByRange(c *gin.Context) {
	minStr := c.DefaultQuery("min", "0")
	maxStr := c.DefaultQuery("max", "100000")

	var (
		min, max int64
		err      error
	)

	_, err = fmt.Sscan(minStr, &min)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid min parameter"})
		return
	}
	_, err = fmt.Sscan(maxStr, &max)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid max parameter"})
		return
	}

	orders, err := h.useCase.GetOrdersByAmountRange(min, max)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, orders)
}

func (h *OrderHandler) ListPayments(c *gin.Context) {
	status := c.Query("status")
	if status == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "status query parameter is required"})
		return
	}

	payments, err := h.useCase.ListPayments(status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, payments)
}
