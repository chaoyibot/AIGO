package response

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// APIResponse is the standard unified success response envelope.
type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

// Meta holds per-request metadata such as request ID and timestamp.
type Meta struct {
	RequestID string `json:"request_id"`
	Timestamp int64  `json:"timestamp"`
}

// Pagination describes cursor-based pagination metadata.
type Pagination struct {
	Cursor  string `json:"cursor"`
	HasMore bool   `json:"has_more"`
	Total   int64  `json:"total,omitempty"`
}

// PaginatedResponse wraps a list of items with pagination metadata.
type PaginatedResponse struct {
	Code       int         `json:"code"`
	Message    string      `json:"message"`
	Data       interface{} `json:"data"`
	Pagination Pagination  `json:"pagination"`
	Meta       Meta        `json:"meta"`
}

// ErrorResponse carries structured error information back to the client.
type ErrorResponse struct {
	Code       int    `json:"code"`
	Message    string `json:"message"`
	Detail     string `json:"detail,omitempty"`
	Suggestion string `json:"suggestion,omitempty"`
	Meta       Meta   `json:"meta"`
}

// Pre-defined application error codes.
const (
	ErrInvalidRequest      = 40000
	ErrUnauthorized        = 40100
	ErrForbidden           = 40300
	ErrNotFound            = 40400
	ErrConflict            = 40900
	ErrInternal            = 50000
	ErrInsufficientPoints  = 40200
	ErrSoldOut             = 41000
)

// Success responds with HTTP 200 and the given data payload.
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, APIResponse{
		Code:    0,
		Message: "success",
		Data:    data,
		Meta: &Meta{
			RequestID: c.GetString("request_id"),
			Timestamp: time.Now().Unix(),
		},
	})
}

// Created responds with HTTP 201 and the given resource data.
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, APIResponse{
		Code:    0,
		Message: "created",
		Data:    data,
		Meta: &Meta{
			RequestID: c.GetString("request_id"),
			Timestamp: time.Now().Unix(),
		},
	})
}

// Error responds with a structured error payload. The httpStatus argument
// sets the HTTP status code while code is the application-level error code.
func Error(c *gin.Context, httpStatus int, code int, message string, detail ...string) {
	d := ""
	if len(detail) > 0 {
		d = detail[0]
	}
	c.JSON(httpStatus, ErrorResponse{
		Code:    code,
		Message: message,
		Detail:  d,
		Meta: Meta{
			RequestID: c.GetString("request_id"),
			Timestamp: time.Now().Unix(),
		},
	})
}

// Paginated responds with HTTP 200 and wraps data with pagination metadata.
func Paginated(c *gin.Context, data interface{}, pagination Pagination) {
	c.JSON(http.StatusOK, PaginatedResponse{
		Code:       0,
		Message:    "success",
		Data:       data,
		Pagination: pagination,
		Meta: Meta{
			RequestID: c.GetString("request_id"),
			Timestamp: time.Now().Unix(),
		},
	})
}
