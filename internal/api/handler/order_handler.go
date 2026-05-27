package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/aigo/internal/api/response"
	"github.com/aigo/internal/repository"
)

type OrderHandler struct {
	orderRepo *repository.OrderRepo
}

func NewOrderHandler(orderRepo *repository.OrderRepo) *OrderHandler {
	return &OrderHandler{orderRepo: orderRepo}
}

func (h *OrderHandler) List(c *gin.Context) {
	userID := c.GetString("user_id")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	orders, err := h.orderRepo.ListByUser(userID, limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.ErrInternal, "list failed")
		return
	}
	response.Success(c, orders)
}

func (h *OrderHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	order, err := h.orderRepo.FindByID(id)
	if err != nil {
		response.Error(c, http.StatusNotFound, response.ErrNotFound, "order not found")
		return
	}
	userID := c.GetString("user_id")
	if order.BuyerID != userID && order.SellerID != userID {
		response.Error(c, http.StatusForbidden, response.ErrForbidden, "access denied")
		return
	}
	response.Success(c, order)
}

func (h *OrderHandler) Confirm(c *gin.Context) {
	id := c.Param("id")
	order, err := h.orderRepo.FindByID(id)
	if err != nil {
		response.Error(c, http.StatusNotFound, response.ErrNotFound, "order not found")
		return
	}
	userID := c.GetString("user_id")
	if order.SellerID != userID {
		response.Error(c, http.StatusForbidden, response.ErrForbidden, "only seller can confirm")
		return
	}
	if err := h.orderRepo.Confirm(id); err != nil {
		response.Error(c, http.StatusInternalServerError, response.ErrInternal, "confirm failed")
		return
	}
	response.Success(c, gin.H{"status": "completed"})
}

func (h *OrderHandler) Cancel(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetString("user_id")
	order, err := h.orderRepo.FindByID(id)
	if err != nil {
		response.Error(c, http.StatusNotFound, response.ErrNotFound, "order not found")
		return
	}
	if order.BuyerID != userID {
		response.Error(c, http.StatusForbidden, response.ErrForbidden, "only buyer can cancel")
		return
	}
	if err := h.orderRepo.Cancel(id); err != nil {
		response.Error(c, http.StatusInternalServerError, response.ErrInternal, "cancel failed")
		return
	}
	response.Success(c, gin.H{"status": "cancelled"})
}

func (h *OrderHandler) Dispute(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetString("user_id")
	var req struct {
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.ErrInvalidRequest, "reason required")
		return
	}
	order, err := h.orderRepo.FindByID(id)
	if err != nil {
		response.Error(c, http.StatusNotFound, response.ErrNotFound, "order not found")
		return
	}
	if order.BuyerID != userID && order.SellerID != userID {
		response.Error(c, http.StatusForbidden, response.ErrForbidden, "access denied")
		return
	}
	if err := h.orderRepo.Dispute(id, req.Reason); err != nil {
		response.Error(c, http.StatusInternalServerError, response.ErrInternal, "dispute failed")
		return
	}
	response.Success(c, gin.H{"status": "disputed"})
}
