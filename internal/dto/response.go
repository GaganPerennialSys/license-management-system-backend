package dto

import (
	"time"
)

// CustomerResponse represents a customer in API responses
type CustomerResponse struct {
	ID        uint      `json:"id"`
	UserID    uint      `json:"user_id"`
	Name      string    `json:"name"`
	Phone     string    `json:"phone"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	User      *UserResponse `json:"user,omitempty"`
}

// UserResponse represents a user in API responses
type UserResponse struct {
	ID        uint      `json:"id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SubscriptionPackResponse represents a subscription pack in API responses
type SubscriptionPackResponse struct {
	ID             uint    `json:"id"`
	Name           string  `json:"name"`
	Description    string  `json:"description"`
	SKU            string  `json:"sku"`
	Price          float64 `json:"price"`
	ValidityMonths int     `json:"validity_months"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// SubscriptionResponse represents a subscription in API responses
type SubscriptionResponse struct {
	ID            uint      `json:"id"`
	CustomerID    uint      `json:"customer_id"`
	PackID        uint      `json:"pack_id"`
	Status        string    `json:"status"`
	RequestedAt   time.Time `json:"requested_at"`
	ApprovedAt    *time.Time `json:"approved_at"`
	AssignedAt    *time.Time `json:"assigned_at"`
	ExpiresAt     *time.Time `json:"expires_at"`
	DeactivatedAt *time.Time `json:"deactivated_at"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	Customer      *CustomerResponse         `json:"customer,omitempty"`
	Pack          *SubscriptionPackResponse `json:"pack,omitempty"`
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

// SuccessResponse represents a successful API response
type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Message string      `json:"message,omitempty"`
}

// ErrorResponse represents an error API response
type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}
