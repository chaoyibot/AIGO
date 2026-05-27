package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/aigo/internal/api/response"
	"github.com/aigo/internal/repository"
	productService "github.com/aigo/internal/service/product"
)

type ProductHandler struct {
	productService *productService.Service
}

func NewProductHandler(productService *productService.Service) *ProductHandler {
	return &ProductHandler{productService: productService}
}

func (h *ProductHandler) Create(c *gin.Context) {
	userID := c.GetString("user_id")
	var req struct {
		EncryptedTitle    []byte   `json:"encrypted_title" binding:"required"`
		EncryptedDesc     []byte   `json:"encrypted_description"`
		EncryptedMetadata []byte   `json:"encrypted_metadata"`
		EncryptedKey      []byte   `json:"encrypted_key" binding:"required"`
		PriceMin          int64    `json:"price_min" binding:"required"`
		PriceMax          int64    `json:"price_max" binding:"required"`
		Category          string   `json:"category"`
		Tags              []string `json:"tags"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.ErrInvalidRequest, "invalid request")
		return
	}

	product, err := h.productService.Create(userID, req.EncryptedTitle, req.EncryptedDesc, req.EncryptedMetadata, req.EncryptedKey, req.PriceMin, req.PriceMax, req.Category, req.Tags)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.ErrInternal, "create failed")
		return
	}
	response.Created(c, gin.H{"product_id": product.ID})
}

func (h *ProductHandler) List(c *gin.Context) {
	category := c.Query("category")
	status := c.Query("status")
	priceMin, _ := strconv.ParseInt(c.Query("price_min"), 10, 64)
	priceMax, _ := strconv.ParseInt(c.Query("price_max"), 10, 64)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	products, cursor, err := h.productService.List(repository.ProductFilter{
		Category: category,
		Status:   status,
		PriceMin: priceMin,
		PriceMax: priceMax,
		Limit:    limit,
	})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, response.ErrInternal, "list failed")
		return
	}

	response.Paginated(c, products, response.Pagination{
		Cursor:  cursor,
		HasMore: cursor != "",
	})
}

func (h *ProductHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	product, err := h.productService.GetByID(id)
	if err != nil {
		response.Error(c, http.StatusNotFound, response.ErrNotFound, "product not found")
		return
	}
	response.Success(c, product)
}

func (h *ProductHandler) Update(c *gin.Context) {
	userID := c.GetString("user_id")
	id := c.Param("id")

	existing, err := h.productService.GetByID(id)
	if err != nil || existing.SellerID != userID {
		response.Error(c, http.StatusNotFound, response.ErrNotFound, "product not found")
		return
	}

	var req struct {
		EncryptedTitle    []byte   `json:"encrypted_title"`
		EncryptedDesc     []byte   `json:"encrypted_description"`
		EncryptedMetadata []byte   `json:"encrypted_metadata"`
		PriceMin          int64    `json:"price_min"`
		PriceMax          int64    `json:"price_max"`
		Category          string   `json:"category"`
		Tags              []string `json:"tags"`
		Status            string   `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, response.ErrInvalidRequest, "invalid request")
		return
	}

	if req.EncryptedTitle != nil { existing.EncryptedTitle = req.EncryptedTitle }
	if req.EncryptedDesc != nil { existing.EncryptedDesc = req.EncryptedDesc }
	if req.EncryptedMetadata != nil { existing.EncryptedMetadata = req.EncryptedMetadata }
	if req.PriceMin > 0 { existing.PriceMin = req.PriceMin }
	if req.PriceMax > 0 { existing.PriceMax = req.PriceMax }
	if req.Category != "" { existing.Category = req.Category }
	if req.Tags != nil { existing.Tags = req.Tags }
	if req.Status != "" { existing.Status = req.Status }

	if err := h.productService.Update(existing); err != nil {
		response.Error(c, http.StatusInternalServerError, response.ErrInternal, "update failed")
		return
	}
	response.Success(c, gin.H{"product_id": id})
}

func (h *ProductHandler) Delete(c *gin.Context) {
	userID := c.GetString("user_id")
	id := c.Param("id")
	if err := h.productService.Delete(id, userID); err != nil {
		response.Error(c, http.StatusInternalServerError, response.ErrInternal, "delete failed")
		return
	}
	response.Success(c, gin.H{"deleted": id})
}
