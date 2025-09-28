package handlers

import (
	"net/http"
	"time"

	"cursor-ai-backend/internal/database"
	"cursor-ai-backend/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type UserHandler struct {
	*BaseHandler
}

func NewUserHandler(db *database.DB) *UserHandler {
	return &UserHandler{
		BaseHandler: NewBaseHandler(db),
	}
}

// LoginRequest represents the login request structure
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// SignupRequest represents the customer signup request structure
type SignupRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name" binding:"required"`
	Phone    string `json:"phone"`
}

// LoginResponse represents the login response structure
type LoginResponse struct {
	Token string      `json:"token"`
	User  models.User `json:"user"`
}

// SDKLoginResponse represents the SDK login response structure
type SDKLoginResponse struct {
	APIKey string      `json:"api_key"`
	User   models.User `json:"user"`
}

// AdminLogin handles admin login
// @Summary Admin login
// @Description Authenticate admin user and return JWT token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login credentials"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/admin/login [post]
func (h *UserHandler) AdminLogin(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.ErrorResponse(c, http.StatusBadRequest, "Invalid request format")
		return
	}

	var user models.User
	err := h.db.Where("email = ? AND role = ?", req.Email, "admin").First(&user).Error
	if err != nil {
		h.ErrorResponse(c, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	if !user.CheckPassword(req.Password) {
		h.ErrorResponse(c, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	token, err := h.generateJWT(&user)
	if err != nil {
		h.ErrorResponse(c, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	// Clear sensitive data
	user.Password = ""
	user.APIKey = nil

	h.SuccessResponse(c, LoginResponse{
		Token: token,
		User:  user,
	}, "Login successful")
}

// CustomerLogin handles customer login
// @Summary Customer login
// @Description Authenticate customer user and return JWT token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login credentials"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/customer/login [post]
func (h *UserHandler) CustomerLogin(c *gin.Context) {
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

	token, err := h.generateJWT(&user)
	if err != nil {
		h.ErrorResponse(c, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	// Clear sensitive data
	user.Password = ""
	user.APIKey = nil

	h.SuccessResponse(c, LoginResponse{
		Token: token,
		User:  user,
	}, "Login successful")
}

// CustomerSignup handles customer registration
// @Summary Customer signup
// @Description Register a new customer account
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body SignupRequest true "Signup information"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /api/customer/signup [post]
func (h *UserHandler) CustomerSignup(c *gin.Context) {
	var req SignupRequest
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

	// Generate API key for customer
	if err := user.GenerateAPIKey(); err != nil {
		h.ErrorResponse(c, http.StatusInternalServerError, "Failed to generate API key")
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

	// Generate JWT token
	token, err := h.generateJWT(user)
	if err != nil {
		h.ErrorResponse(c, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	// Load customer relationship
	h.db.Preload("Customer").First(user, user.ID)

	// Clear sensitive data
	user.Password = ""
	user.APIKey = nil

	h.SuccessResponse(c, LoginResponse{
		Token: token,
		User:  *user,
	}, "Registration successful")
}

// generateJWT creates a JWT token for the user
func (h *UserHandler) generateJWT(user *models.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"role":    user.Role,
		"exp":     time.Now().Add(time.Hour * 24).Unix(), // 24 hours
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte("your-secret-key-change-in-production")) // TODO: Use config
}
