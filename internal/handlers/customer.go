package handlers

import (
	"net/http"
	"strconv"

	"cursor-ai-backend/internal/database"
	"cursor-ai-backend/internal/models"

	"github.com/gin-gonic/gin"
)

type CustomerHandler struct {
	*BaseHandler
}

func NewCustomerHandler(db *database.DB) *CustomerHandler {
	return &CustomerHandler{
		BaseHandler: NewBaseHandler(db),
	}
}

// CreateCustomerRequest represents the customer creation request
type CreateCustomerRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name" binding:"required"`
	Phone    string `json:"phone"`
}

// UpdateCustomerRequest represents the customer update request
type UpdateCustomerRequest struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
}

// ListCustomers handles listing all customers (admin only)
// @Summary List customers
// @Description Get paginated list of all customers
// @Tags Admin Customer Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param search query string false "Search term"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /api/v1/admin/customers [get]
func (h *CustomerHandler) ListCustomers(c *gin.Context) {
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
	query := h.db.Preload("User").Model(&models.Customer{})

	// Apply search filter
	if search != "" {
		query = query.Joins("JOIN users ON customers.user_id = users.id").
			Where("customers.name ILIKE ? OR users.email ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// Get total count
	var total int64
	query.Count(&total)

	// Get customers
	var customers []models.Customer
	err := query.Offset(offset).Limit(limit).Find(&customers).Error
	if err != nil {
		h.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve customers")
		return
	}

	h.PaginatedResponse(c, customers, total, page, limit)
}

// CreateCustomer handles creating a new customer (admin only)
// @Summary Create customer
// @Description Create a new customer account
// @Tags Admin Customer Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateCustomerRequest true "Customer information"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /api/v1/admin/customers [post]
func (h *CustomerHandler) CreateCustomer(c *gin.Context) {
	var req CreateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.ErrorResponse(c, http.StatusBadRequest, "Invalid request format")
		return
	}

	// Check if user already exists
	var existingUser models.User
	err := h.db.Where("email = ?", req.Email).First(&existingUser).Error
	if err == nil {
		h.ErrorResponse(c, http.StatusConflict, "Email already registered")
		return
	}

	// Create user
	user := &models.User{
		Email:    req.Email,
		Password: req.Password,
		Role:     "customer",
	}

	if err := user.HashPassword(); err != nil {
		h.ErrorResponse(c, http.StatusInternalServerError, "Failed to process password")
		return
	}

	if err := h.db.Create(user).Error; err != nil {
		h.ErrorResponse(c, http.StatusInternalServerError, "Failed to create user")
		return
	}

	// Create customer profile
	customer := &models.Customer{
		UserID: user.ID,
		Name:   req.Name,
		Phone:  req.Phone,
	}

	if err := h.db.Create(customer).Error; err != nil {
		h.ErrorResponse(c, http.StatusInternalServerError, "Failed to create customer profile")
		return
	}

	// Load user relationship
	h.db.Preload("User").First(customer, customer.ID)

	h.SuccessResponse(c, customer, "Customer created successfully")
}

// GetCustomer handles getting a specific customer (admin only)
// @Summary Get customer
// @Description Get customer details by ID
// @Tags Admin Customer Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Customer ID"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/admin/customers/{id} [get]
func (h *CustomerHandler) GetCustomer(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.ErrorResponse(c, http.StatusBadRequest, "Invalid customer ID")
		return
	}

	var customer models.Customer
	err = h.db.Preload("User").Preload("Subscriptions.Pack").First(&customer, id).Error
	if err != nil {
		h.ErrorResponse(c, http.StatusNotFound, "Customer not found")
		return
	}

	h.SuccessResponse(c, customer, "Customer retrieved successfully")
}

// UpdateCustomer handles updating a customer (admin only)
// @Summary Update customer
// @Description Update customer information
// @Tags Admin Customer Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Customer ID"
// @Param request body UpdateCustomerRequest true "Updated customer information"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/admin/customers/{id} [put]
func (h *CustomerHandler) UpdateCustomer(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.ErrorResponse(c, http.StatusBadRequest, "Invalid customer ID")
		return
	}

	var req UpdateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.ErrorResponse(c, http.StatusBadRequest, "Invalid request format")
		return
	}

	var customer models.Customer
	err = h.db.First(&customer, id).Error
	if err != nil {
		h.ErrorResponse(c, http.StatusNotFound, "Customer not found")
		return
	}

	// Update fields
	if req.Name != "" {
		customer.Name = req.Name
	}
	if req.Phone != "" {
		customer.Phone = req.Phone
	}

	if err := h.db.Save(&customer).Error; err != nil {
		h.ErrorResponse(c, http.StatusInternalServerError, "Failed to update customer")
		return
	}

	// Load user relationship
	h.db.Preload("User").First(&customer, customer.ID)

	h.SuccessResponse(c, customer, "Customer updated successfully")
}

// DeleteCustomer handles soft deleting a customer (admin only)
// @Summary Delete customer
// @Description Soft delete a customer account
// @Tags Admin Customer Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Customer ID"
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/admin/customers/{id} [delete]
func (h *CustomerHandler) DeleteCustomer(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.ErrorResponse(c, http.StatusBadRequest, "Invalid customer ID")
		return
	}

	var customer models.Customer
	err = h.db.First(&customer, id).Error
	if err != nil {
		h.ErrorResponse(c, http.StatusNotFound, "Customer not found")
		return
	}

	if err := h.db.Delete(&customer).Error; err != nil {
		h.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete customer")
		return
	}

	h.SuccessResponse(c, gin.H{"message": "Customer deleted successfully"}, "")
}

// GetProfile handles getting current customer's profile
// @Summary Get profile
// @Description Get current customer's profile information
// @Tags Customer Profile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /api/v1/customer/profile [get]
func (h *CustomerHandler) GetProfile(c *gin.Context) {
	customer, err := h.GetCurrentCustomer(c)
	if err != nil {
		h.ErrorResponse(c, http.StatusUnauthorized, "Customer not found")
		return
	}

	// Load user relationship
	h.db.Preload("User").Preload("Subscriptions.Pack").First(customer, customer.ID)

	h.SuccessResponse(c, customer, "Profile retrieved successfully")
}

// UpdateProfile handles updating current customer's profile
// @Summary Update profile
// @Description Update current customer's profile information
// @Tags Customer Profile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body UpdateCustomerRequest true "Updated profile information"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /api/v1/customer/profile [put]
func (h *CustomerHandler) UpdateProfile(c *gin.Context) {
	customer, err := h.GetCurrentCustomer(c)
	if err != nil {
		h.ErrorResponse(c, http.StatusUnauthorized, "Customer not found")
		return
	}

	var req UpdateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.ErrorResponse(c, http.StatusBadRequest, "Invalid request format")
		return
	}

	// Update fields
	if req.Name != "" {
		customer.Name = req.Name
	}
	if req.Phone != "" {
		customer.Phone = req.Phone
	}

	if err := h.db.Save(customer).Error; err != nil {
		h.ErrorResponse(c, http.StatusInternalServerError, "Failed to update profile")
		return
	}

	// Load user relationship
	h.db.Preload("User").First(customer, customer.ID)

	h.SuccessResponse(c, customer, "Profile updated successfully")
}
