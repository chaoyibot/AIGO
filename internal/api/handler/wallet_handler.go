package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/aigo/internal/api/response"
	walletService "github.com/aigo/internal/service/wallet"
)

type WalletHandler struct {
	walletService *walletService.Service
}

func NewWalletHandler(walletService *walletService.Service) *WalletHandler {
	return &WalletHandler{walletService: walletService}
}

func (h *WalletHandler) GetBalance(c *gin.Context) {
	userID := c.GetString("user_id")
	wallet, err := h.walletService.GetBalance(userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.ErrInternal, "get balance failed")
		return
	}
	response.Success(c, gin.H{
		"fiat_balance":   wallet.FiatBalance,
		"points_balance": wallet.PointsBalance,
		"currency":       wallet.Currency,
	})
}

func (h *WalletHandler) Recharge(c *gin.Context) {
	userID := c.GetString("user_id")
	var req struct {
		Amount int64 `json:"amount" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Amount <= 0 {
		response.Error(c, http.StatusBadRequest, response.ErrInvalidRequest, "amount must be positive")
		return
	}

	order, err := h.walletService.Recharge(userID, req.Amount)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.ErrInternal, "recharge failed")
		return
	}
	response.Created(c, gin.H{
		"order_id":       order.ID,
		"amount":         order.Amount,
		"points_awarded": order.PointsAwarded,
		"status":         order.Status,
	})
}

func (h *WalletHandler) Convert(c *gin.Context) {
	userID := c.GetString("user_id")
	var req struct {
		FiatAmount int64 `json:"fiat_amount" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.FiatAmount <= 0 {
		response.Error(c, http.StatusBadRequest, response.ErrInvalidRequest, "invalid amount")
		return
	}

	if err := h.walletService.ConvertToPoints(userID, req.FiatAmount); err != nil {
		response.Error(c, http.StatusBadRequest, response.ErrInsufficientPoints, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "conversion successful"})
}

func (h *WalletHandler) GetTransactions(c *gin.Context) {
	userID := c.GetString("user_id")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	txns, err := h.walletService.GetTransactions(userID, limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.ErrInternal, "query failed")
		return
	}
	response.Success(c, txns)
}
