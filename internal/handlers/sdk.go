package handlers

import (
	"net/http"
	"strconv"
	"time"

	"cursor-ai-backend/internal/database"
	"cursor-ai-backend/internal/models"

	"github.com/gin-gonic/gin"
)

type SDKHandler struct {
	*BaseHandler
}

func NewSDKHandler(db *database.DB) *SDKHandler {
	return &SDKHandler{
		BaseHandler: NewBaseHandler(db),
	}
}

// SDKLogin handles SDK authentication and returns API key
// @Summary SDK login
// @Description Authenticate user for SDK access and return API key
// @Tags SDK Authentication
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login credentials"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /sdk/auth/login [post]
func (h *SDKHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.ErrorResponse(c, http.StatusBadRequest, "Invalid request format")
		return
	}

	var user models.User
	err := h.db.Preload("Customer").Where("email = ? AND role = ?", req.Email, "customer").First(&user).Error
	if err != nil {
		h.ErrorResponse(c, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	if !user.CheckPassword(req.Password) {
		h.ErrorResponse(c, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// Generate API key if user doesn't have one
	if !user.HasAPIKey() {
		if err := user.GenerateAPIKey(); err != nil {
			h.ErrorResponse(c, http.StatusInternalServerError, "Failed to generate API key")
			return
		}

		if err := h.db.Save(&user).Error; err != nil {
			h.ErrorResponse(c, http.StatusInternalServerError, "Failed to save API key")
			return
		}
	}

	// Clear sensitive data
	user.Password = ""

	apiKey := ""
	if user.APIKey != nil {
		apiKey = *user.APIKey
	}

	h.SuccessResponse(c, SDKLoginResponse{
		APIKey: apiKey,
		User:   user,
	}, "SDK authentication successful")
}

// GetCurrentSubscription returns the customer's current active subscription
// @Summary Get current subscription
// @Description Get the customer's currently active subscription
// @Tags SDK Subscription
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /sdk/v1/subscription [get]
func (h *SDKHandler) GetCurrentSubscription(c *gin.Context) {
	customer, err := h.GetCurrentCustomer(c)
	if err != nil {
		h.ErrorResponse(c, http.StatusUnauthorized, "Customer not found")
		return
	}

	subscription, err := customer.GetActiveSubscription(h.db.DB)
	if err != nil {
		h.ErrorResponse(c, http.StatusNotFound, "No active subscription found")
		return
	}

	// Load pack information
	h.db.Preload("Pack").First(subscription, subscription.ID)

	h.SuccessResponse(c, subscription, "Current subscription retrieved")
}

// RequestSubscription allows customer to request a new subscription
// @Summary Request subscription
// @Description Request a new subscription for the customer
// @Tags SDK Subscription
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body SubscriptionRequest true "Subscription request"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /sdk/v1/subscription/request [post]
func (h *SDKHandler) RequestSubscription(c *gin.Context) {
	customer, err := h.GetCurrentCustomer(c)
	if err != nil {
		h.ErrorResponse(c, http.StatusUnauthorized, "Customer not found")
		return
	}

	var req SubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.ErrorResponse(c, http.StatusBadRequest, "Invalid request format")
		return
	}

	// Check if customer already has an active subscription
	if customer.HasActiveSubscription(h.db.DB) {
		h.ErrorResponse(c, http.StatusConflict, "Customer already has an active subscription")
		return
	}

	// Verify subscription pack exists
	var pack models.SubscriptionPack
	err = h.db.Where("sku = ?", req.PackSKU).First(&pack).Error
	if err != nil {
		h.ErrorResponse(c, http.StatusBadRequest, "Invalid subscription pack")
		return
	}

	// Create subscription request
	subscription := &models.Subscription{
		CustomerID:  customer.ID,
		PackID:      pack.ID,
		Status:      models.StatusRequested,
		RequestedAt: time.Now(),
	}

	if err := h.db.Create(subscription).Error; err != nil {
		h.ErrorResponse(c, http.StatusInternalServerError, "Failed to create subscription request")
		return
	}

	// Load pack information
	h.db.Preload("Pack").First(subscription, subscription.ID)

	h.SuccessResponse(c, subscription, "Subscription request created successfully")
}

// DeactivateSubscription allows customer to deactivate their current subscription
// @Summary Deactivate subscription
// @Description Deactivate the customer's current active subscription
// @Tags SDK Subscription
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /sdk/v1/subscription/deactivate [put]
func (h *SDKHandler) DeactivateSubscription(c *gin.Context) {
	customer, err := h.GetCurrentCustomer(c)
	if err != nil {
		h.ErrorResponse(c, http.StatusUnauthorized, "Customer not found")
		return
	}

	subscription, err := customer.GetActiveSubscription(h.db.DB)
	if err != nil {
		h.ErrorResponse(c, http.StatusNotFound, "No active subscription found")
		return
	}

	// Update subscription status
	subscription.Status = models.StatusInactive
	now := time.Now()
	subscription.DeactivatedAt = &now

	if err := h.db.Save(subscription).Error; err != nil {
		h.ErrorResponse(c, http.StatusInternalServerError, "Failed to deactivate subscription")
		return
	}

	// Load pack information
	h.db.Preload("Pack").First(subscription, subscription.ID)

	h.SuccessResponse(c, subscription, "Subscription deactivated successfully")
}

// GetSubscriptionHistory returns paginated subscription history for the customer
// @Summary Get subscription history
// @Description Get paginated history of customer's subscriptions
// @Tags SDK Subscription
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param sort query string false "Sort field" default(created_at)
// @Param order query string false "Sort order" default(desc)
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Router /sdk/v1/subscription/history [get]
func (h *SDKHandler) GetSubscriptionHistory(c *gin.Context) {
	customer, err := h.GetCurrentCustomer(c)
	if err != nil {
		h.ErrorResponse(c, http.StatusUnauthorized, "Customer not found")
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	sort := c.DefaultQuery("sort", "created_at")
	order := c.DefaultQuery("order", "desc")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	// Build query
	query := h.db.Preload("Pack").Where("customer_id = ?", customer.ID)

	// Apply sorting
	if order == "asc" {
		query = query.Order(sort + " ASC")
	} else {
		query = query.Order(sort + " DESC")
	}

	// Get total count
	var total int64
	h.db.Model(&models.Subscription{}).Where("customer_id = ?", customer.ID).Count(&total)

	// Get subscriptions
	var subscriptions []models.Subscription
	err = query.Offset(offset).Limit(limit).Find(&subscriptions).Error
	if err != nil {
		h.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve subscription history")
		return
	}

	h.PaginatedResponse(c, subscriptions, total, page, limit)
}
