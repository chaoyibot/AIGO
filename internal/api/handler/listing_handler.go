package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/aigo/internal/api/response"
	"github.com/aigo/internal/model"
	"github.com/aigo/internal/repository"
	tradingService "github.com/aigo/internal/service/trading"
)

type ListingHandler struct {
	listingRepo *repository.ListingRepo
	tradingSvc  *tradingService.Service
}

func NewListingHandler(listingRepo *repository.ListingRepo, tradingSvc *tradingService.Service) *ListingHandler {
	return &ListingHandler{listingRepo: listingRepo, tradingSvc: tradingSvc}
}

func (h *ListingHandler) Create(c *gin.Context) {
	userID := c.GetString("user_id")

	var req struct {
		ProductID string `json:"product_id" binding:"required"`
		PriceType string `json:"price_type"`
		Price     int64  `json:"price" binding:"required"`
		Quantity  int    `json:"quantity"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.ErrInvalidRequest, "invalid request")
		return
	}
	if req.PriceType == "" {
		req.PriceType = "fixed"
	}
	if req.Quantity <= 0 {
		req.Quantity = 1
	}

	listing := &model.Listing{
		ID:        uuid.New().String(),
		ProductID: req.ProductID,
		SellerID:  userID,
		PriceType: req.PriceType,
		Price:     req.Price,
		Quantity:  req.Quantity,
		Status:    "active",
	}

	if err := h.listingRepo.Create(listing); err != nil {
		response.Error(c, http.StatusInternalServerError, response.ErrInternal, "create listing failed")
		return
	}

	response.Created(c, gin.H{"listing_id": listing.ID})
}

func (h *ListingHandler) List(c *gin.Context) {
	limit := 20
	listings, err := h.listingRepo.ListActive(limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.ErrInternal, "list failed")
		return
	}
	if listings == nil {
		listings = []model.Listing{}
	}
	response.Success(c, listings)
}

func (h *ListingHandler) Buy(c *gin.Context) {
	userID := c.GetString("user_id")
	listingID := c.Param("id")

	var req struct {
		Quantity int `json:"quantity"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Quantity = 1
	}

	order, err := h.tradingSvc.BuyListing(userID, listingID, req.Quantity)
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.ErrConflict, err.Error())
		return
	}

	response.Created(c, gin.H{
		"order_id":    order.ID,
		"total_price": order.TotalPrice,
		"status":      order.Status,
	})
}
