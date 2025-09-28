package handlers

import (
	"fmt"

	"cursor-ai-backend/internal/database"
	"cursor-ai-backend/internal/models"

	"github.com/gin-gonic/gin"
)

type BaseHandler struct {
	db *database.DB
}

func NewBaseHandler(db *database.DB) *BaseHandler {
	return &BaseHandler{db: db}
}

// GetCurrentUser retrieves the current user from the context
func (h *BaseHandler) GetCurrentUser(c *gin.Context) (*models.User, error) {
	userID, exists := c.Get("user_id")
	if !exists {
		return nil, fmt.Errorf("unauthorized")
	}

	var user models.User
	err := h.db.Preload("Customer").First(&user, userID).Error
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GetCurrentCustomer retrieves the current customer from the context
func (h *BaseHandler) GetCurrentCustomer(c *gin.Context) (*models.Customer, error) {
	user, err := h.GetCurrentUser(c)
	if err != nil {
		return nil, err
	}

	if user.Customer == nil {
		return nil, fmt.Errorf("customer not found")
	}

	return user.Customer, nil
}

// SuccessResponse creates a standardized success response
func (h *BaseHandler) SuccessResponse(c *gin.Context, data interface{}, message string) {
	response := gin.H{
		"success": true,
		"data":    data,
	}
	if message != "" {
		response["message"] = message
	}
	c.JSON(200, response)
}

// ErrorResponse creates a standardized error response
func (h *BaseHandler) ErrorResponse(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, gin.H{
		"success": false,
		"error":   message,
	})
}

// PaginatedResponse creates a standardized paginated response
func (h *BaseHandler) PaginatedResponse(c *gin.Context, data interface{}, total int64, page, limit int) {
	response := gin.H{
		"success": true,
		"data":    data,
		"pagination": gin.H{
			"total": total,
			"page":  page,
			"limit": limit,
		},
	}
	c.JSON(200, response)
}

// SubscriptionRequest represents a subscription request
type SubscriptionRequest struct {
	PackSKU string `json:"pack_sku" binding:"required"`
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	Success    bool        `json:"success"`
	Data       interface{} `json:"data"`
	Pagination struct {
		Total int64 `json:"total"`
		Page  int   `json:"page"`
		Limit int   `json:"limit"`
	} `json:"pagination"`
}
