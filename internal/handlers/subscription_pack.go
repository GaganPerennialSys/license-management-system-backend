package handlers

import (
	"net/http"
	"strconv"

	"cursor-ai-backend/internal/database"
	"cursor-ai-backend/internal/models"

	"github.com/gin-gonic/gin"
)

type SubscriptionPackHandler struct {
	*BaseHandler
}

func NewSubscriptionPackHandler(db *database.DB) *SubscriptionPackHandler {
	return &SubscriptionPackHandler{
		BaseHandler: NewBaseHandler(db),
	}
}

// CreatePackRequest represents the subscription pack creation request
type CreatePackRequest struct {
	Name           string  `json:"name" binding:"required"`
	Description    string  `json:"description"`
	SKU            string  `json:"sku" binding:"required"`
	Price          float64 `json:"price" binding:"required,min=0"`
	ValidityMonths int     `json:"validity_months" binding:"required,min=1,max=12"`
}

// UpdatePackRequest represents the subscription pack update request
type UpdatePackRequest struct {
	Name           string  `json:"name"`
	Description    string  `json:"description"`
	Price          float64 `json:"price" binding:"min=0"`
	ValidityMonths int     `json:"validity_months" binding:"min=1,max=12"`
}

// ListPacks handles listing all subscription packs (admin only)
// @Summary List subscription packs
// @Description Get paginated list of all subscription packs
// @Tags Admin Subscription Pack Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param search query string false "Search term"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /api/v1/admin/packs [get]
func (h *SubscriptionPackHandler) ListPacks(c *gin.Context) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	search := c.Query("search")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	// Build query
	query := h.db.Model(&models.SubscriptionPack{})

	// Apply search filter
	if search != "" {
		query = query.Where("name ILIKE ? OR description ILIKE ? OR sku ILIKE ?", 
			"%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	// Get total count
	var total int64
	query.Count(&total)

	// Get packs
	var packs []models.SubscriptionPack
	err := query.Offset(offset).Limit(limit).Find(&packs).Error
	if err != nil {
		h.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve subscription packs")
		return
	}

	h.PaginatedResponse(c, packs, total, page, limit)
}

// CreatePack handles creating a new subscription pack (admin only)
// @Summary Create subscription pack
// @Description Create a new subscription pack
// @Tags Admin Subscription Pack Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreatePackRequest true "Subscription pack information"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /api/v1/admin/packs [post]
func (h *SubscriptionPackHandler) CreatePack(c *gin.Context) {
	var req CreatePackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.ErrorResponse(c, http.StatusBadRequest, "Invalid request format")
		return
	}

	// Check if SKU already exists
	var existingPack models.SubscriptionPack
	err := h.db.Where("sku = ?", req.SKU).First(&existingPack).Error
	if err == nil {
		h.ErrorResponse(c, http.StatusConflict, "SKU already exists")
		return
	}

	// Create subscription pack
	pack := &models.SubscriptionPack{
		Name:           req.Name,
		Description:    req.Description,
		SKU:            req.SKU,
		Price:          req.Price,
		ValidityMonths: req.ValidityMonths,
	}

	if err := h.db.Create(pack).Error; err != nil {
		h.ErrorResponse(c, http.StatusInternalServerError, "Failed to create subscription pack")
		return
	}

	h.SuccessResponse(c, pack, "Subscription pack created successfully")
}

// GetPack handles getting a specific subscription pack (admin only)
// @Summary Get subscription pack
// @Description Get subscription pack details by ID
// @Tags Admin Subscription Pack Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Subscription Pack ID"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/admin/packs/{id} [get]
func (h *SubscriptionPackHandler) GetPack(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.ErrorResponse(c, http.StatusBadRequest, "Invalid subscription pack ID")
		return
	}

	var pack models.SubscriptionPack
	err = h.db.Preload("Subscriptions.Customer").First(&pack, id).Error
	if err != nil {
		h.ErrorResponse(c, http.StatusNotFound, "Subscription pack not found")
		return
	}

	h.SuccessResponse(c, pack, "Subscription pack retrieved successfully")
}

// UpdatePack handles updating a subscription pack (admin only)
// @Summary Update subscription pack
// @Description Update subscription pack information
// @Tags Admin Subscription Pack Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Subscription Pack ID"
// @Param request body UpdatePackRequest true "Updated subscription pack information"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/admin/packs/{id} [put]
func (h *SubscriptionPackHandler) UpdatePack(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.ErrorResponse(c, http.StatusBadRequest, "Invalid subscription pack ID")
		return
	}

	var req UpdatePackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.ErrorResponse(c, http.StatusBadRequest, "Invalid request format")
		return
	}

	var pack models.SubscriptionPack
	err = h.db.First(&pack, id).Error
	if err != nil {
		h.ErrorResponse(c, http.StatusNotFound, "Subscription pack not found")
		return
	}

	// Update fields
	if req.Name != "" {
		pack.Name = req.Name
	}
	if req.Description != "" {
		pack.Description = req.Description
	}
	if req.Price > 0 {
		pack.Price = req.Price
	}
	if req.ValidityMonths > 0 {
		pack.ValidityMonths = req.ValidityMonths
	}

	if err := h.db.Save(&pack).Error; err != nil {
		h.ErrorResponse(c, http.StatusInternalServerError, "Failed to update subscription pack")
		return
	}

	h.SuccessResponse(c, pack, "Subscription pack updated successfully")
}

// DeletePack handles soft deleting a subscription pack (admin only)
// @Summary Delete subscription pack
// @Description Soft delete a subscription pack
// @Tags Admin Subscription Pack Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Subscription Pack ID"
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/admin/packs/{id} [delete]
func (h *SubscriptionPackHandler) DeletePack(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.ErrorResponse(c, http.StatusBadRequest, "Invalid subscription pack ID")
		return
	}

	var pack models.SubscriptionPack
	err = h.db.First(&pack, id).Error
	if err != nil {
		h.ErrorResponse(c, http.StatusNotFound, "Subscription pack not found")
		return
	}

	if err := h.db.Delete(&pack).Error; err != nil {
		h.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete subscription pack")
		return
	}

	h.SuccessResponse(c, gin.H{"message": "Subscription pack deleted successfully"}, "")
}
