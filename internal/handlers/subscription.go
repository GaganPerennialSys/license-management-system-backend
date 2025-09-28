package handlers

import (
	"net/http"
	"strconv"
	"time"

	"cursor-ai-backend/internal/database"
	"cursor-ai-backend/internal/models"

	"github.com/gin-gonic/gin"
)

type SubscriptionHandler struct {
	*BaseHandler
}

func NewSubscriptionHandler(db *database.DB) *SubscriptionHandler {
	return &SubscriptionHandler{
		BaseHandler: NewBaseHandler(db),
	}
}

// CreateSubscriptionRequest represents the subscription creation request (admin only)
type CreateSubscriptionRequest struct {
	CustomerID uint   `json:"customer_id" binding:"required"`
	PackSKU    string `json:"pack_sku" binding:"required"`
}


// ListSubscriptions handles listing all subscriptions (admin only)
// @Summary List subscriptions
// @Description Get paginated list of all subscriptions
// @Tags Admin Subscription Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param status query string false "Filter by status"
// @Param customer_id query int false "Filter by customer ID"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /api/v1/admin/subscriptions [get]
func (h *SubscriptionHandler) ListSubscriptions(c *gin.Context) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	status := c.Query("status")
	customerIDStr := c.Query("customer_id")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	// Build query
	query := h.db.Preload("Customer.User").Preload("Pack").Model(&models.Subscription{})

	// Apply filters
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if customerIDStr != "" {
		if customerID, err := strconv.ParseUint(customerIDStr, 10, 32); err == nil {
			query = query.Where("customer_id = ?", customerID)
		}
	}

	// Get total count
	var total int64
	query.Count(&total)

	// Get subscriptions
	var subscriptions []models.Subscription
	err := query.Offset(offset).Limit(limit).Find(&subscriptions).Error
	if err != nil {
		h.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve subscriptions")
		return
	}

	h.PaginatedResponse(c, subscriptions, total, page, limit)
}

// CreateSubscription handles creating a new subscription (admin only)
// @Summary Create subscription
// @Description Create a new subscription for a customer
// @Tags Admin Subscription Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateSubscriptionRequest true "Subscription information"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /api/v1/admin/subscriptions [post]
func (h *SubscriptionHandler) CreateSubscription(c *gin.Context) {
	var req CreateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.ErrorResponse(c, http.StatusBadRequest, "Invalid request format")
		return
	}

	// Verify customer exists
	var customer models.Customer
	err := h.db.First(&customer, req.CustomerID).Error
	if err != nil {
		h.ErrorResponse(c, http.StatusNotFound, "Customer not found")
		return
	}

	// Verify subscription pack exists
	var pack models.SubscriptionPack
	err = h.db.Where("sku = ?", req.PackSKU).First(&pack).Error
	if err != nil {
		h.ErrorResponse(c, http.StatusNotFound, "Subscription pack not found")
		return
	}

	// Check if customer already has an active subscription
	if customer.HasActiveSubscription(h.db.DB) {
		h.ErrorResponse(c, http.StatusConflict, "Customer already has an active subscription")
		return
	}

	// Create subscription
	subscription := &models.Subscription{
		CustomerID:  customer.ID,
		PackID:      pack.ID,
		Status:      models.StatusRequested,
		RequestedAt: time.Now(),
	}

	if err := h.db.Create(subscription).Error; err != nil {
		h.ErrorResponse(c, http.StatusInternalServerError, "Failed to create subscription")
		return
	}

	// Load relationships
	h.db.Preload("Customer.User").Preload("Pack").First(subscription, subscription.ID)

	h.SuccessResponse(c, subscription, "Subscription created successfully")
}

// GetSubscription handles getting a specific subscription (admin only)
// @Summary Get subscription
// @Description Get subscription details by ID
// @Tags Admin Subscription Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Subscription ID"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/admin/subscriptions/{id} [get]
func (h *SubscriptionHandler) GetSubscription(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.ErrorResponse(c, http.StatusBadRequest, "Invalid subscription ID")
		return
	}

	var subscription models.Subscription
	err = h.db.Preload("Customer.User").Preload("Pack").First(&subscription, id).Error
	if err != nil {
		h.ErrorResponse(c, http.StatusNotFound, "Subscription not found")
		return
	}

	h.SuccessResponse(c, subscription, "Subscription retrieved successfully")
}

// ApproveSubscription handles approving a subscription request (admin only)
// @Summary Approve subscription
// @Description Approve a subscription request
// @Tags Admin Subscription Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Subscription ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/admin/subscriptions/{id}/approve [put]
func (h *SubscriptionHandler) ApproveSubscription(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.ErrorResponse(c, http.StatusBadRequest, "Invalid subscription ID")
		return
	}

	var subscription models.Subscription
	err = h.db.Preload("Pack").First(&subscription, id).Error
	if err != nil {
		h.ErrorResponse(c, http.StatusNotFound, "Subscription not found")
		return
	}

	// Check if subscription can be approved
	if !subscription.CanTransitionTo(models.StatusApproved) {
		h.ErrorResponse(c, http.StatusBadRequest, "Subscription cannot be approved in current status")
		return
	}

	// Update subscription status
	subscription.Status = models.StatusApproved
	now := time.Now()
	subscription.ApprovedAt = &now

	if err := h.db.Save(&subscription).Error; err != nil {
		h.ErrorResponse(c, http.StatusInternalServerError, "Failed to approve subscription")
		return
	}

	// Load relationships
	h.db.Preload("Customer.User").Preload("Pack").First(&subscription, subscription.ID)

	h.SuccessResponse(c, subscription, "Subscription approved successfully")
}

// AssignSubscription handles assigning an approved subscription (admin only)
// @Summary Assign subscription
// @Description Assign an approved subscription to make it active
// @Tags Admin Subscription Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Subscription ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/admin/subscriptions/{id}/assign [put]
func (h *SubscriptionHandler) AssignSubscription(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.ErrorResponse(c, http.StatusBadRequest, "Invalid subscription ID")
		return
	}

	var subscription models.Subscription
	err = h.db.Preload("Pack").First(&subscription, id).Error
	if err != nil {
		h.ErrorResponse(c, http.StatusNotFound, "Subscription not found")
		return
	}

	// Check if subscription can be assigned
	if !subscription.CanTransitionTo(models.StatusActive) {
		h.ErrorResponse(c, http.StatusBadRequest, "Subscription cannot be assigned in current status")
		return
	}

	// Check if customer already has an active subscription
	var customer models.Customer
	err = h.db.First(&customer, subscription.CustomerID).Error
	if err != nil {
		h.ErrorResponse(c, http.StatusNotFound, "Customer not found")
		return
	}

	if customer.HasActiveSubscription(h.db.DB) {
		h.ErrorResponse(c, http.StatusConflict, "Customer already has an active subscription")
		return
	}

	// Update subscription status
	subscription.Status = models.StatusActive
	now := time.Now()
	subscription.AssignedAt = &now
	subscription.CalculateExpiry(subscription.Pack)

	if err := h.db.Save(&subscription).Error; err != nil {
		h.ErrorResponse(c, http.StatusInternalServerError, "Failed to assign subscription")
		return
	}

	// Load relationships
	h.db.Preload("Customer.User").Preload("Pack").First(&subscription, subscription.ID)

	h.SuccessResponse(c, subscription, "Subscription assigned successfully")
}

// UnassignSubscription handles unassigning an active subscription (admin only)
// @Summary Unassign subscription
// @Description Unassign an active subscription
// @Tags Admin Subscription Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Subscription ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/admin/subscriptions/{id}/unassign [put]
func (h *SubscriptionHandler) UnassignSubscription(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.ErrorResponse(c, http.StatusBadRequest, "Invalid subscription ID")
		return
	}

	var subscription models.Subscription
	err = h.db.First(&subscription, id).Error
	if err != nil {
		h.ErrorResponse(c, http.StatusNotFound, "Subscription not found")
		return
	}

	// Check if subscription can be unassigned
	if subscription.Status != models.StatusActive {
		h.ErrorResponse(c, http.StatusBadRequest, "Only active subscriptions can be unassigned")
		return
	}

	// Update subscription status
	subscription.Status = models.StatusInactive
	now := time.Now()
	subscription.DeactivatedAt = &now

	if err := h.db.Save(&subscription).Error; err != nil {
		h.ErrorResponse(c, http.StatusInternalServerError, "Failed to unassign subscription")
		return
	}

	// Load relationships
	h.db.Preload("Customer.User").Preload("Pack").First(&subscription, subscription.ID)

	h.SuccessResponse(c, subscription, "Subscription unassigned successfully")
}

// DeleteSubscription handles deleting a subscription (admin only)
// @Summary Delete subscription
// @Description Delete a subscription
// @Tags Admin Subscription Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Subscription ID"
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/admin/subscriptions/{id} [delete]
func (h *SubscriptionHandler) DeleteSubscription(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.ErrorResponse(c, http.StatusBadRequest, "Invalid subscription ID")
		return
	}

	var subscription models.Subscription
	err = h.db.First(&subscription, id).Error
	if err != nil {
		h.ErrorResponse(c, http.StatusNotFound, "Subscription not found")
		return
	}

	if err := h.db.Delete(&subscription).Error; err != nil {
		h.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete subscription")
		return
	}

	h.SuccessResponse(c, gin.H{"message": "Subscription deleted successfully"}, "")
}

// GetCurrentSubscription handles getting current customer's active subscription
// @Summary Get current subscription
// @Description Get current customer's active subscription
// @Tags Customer Subscription
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/customer/subscription [get]
func (h *SubscriptionHandler) GetCurrentSubscription(c *gin.Context) {
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

// RequestSubscription handles customer subscription request
// @Summary Request subscription
// @Description Request a new subscription for current customer
// @Tags Customer Subscription
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body SubscriptionRequest true "Subscription request"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /api/v1/customer/subscription/request [post]
func (h *SubscriptionHandler) RequestSubscription(c *gin.Context) {
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

// DeactivateSubscription handles customer subscription deactivation
// @Summary Deactivate subscription
// @Description Deactivate current customer's active subscription
// @Tags Customer Subscription
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/customer/subscription/deactivate [put]
func (h *SubscriptionHandler) DeactivateSubscription(c *gin.Context) {
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

// GetSubscriptionHistory handles getting customer's subscription history
// @Summary Get subscription history
// @Description Get paginated history of current customer's subscriptions
// @Tags Customer Subscription
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param sort query string false "Sort field" default(created_at)
// @Param order query string false "Sort order" default(desc)
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /api/v1/customer/subscription/history [get]
func (h *SubscriptionHandler) GetSubscriptionHistory(c *gin.Context) {
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
