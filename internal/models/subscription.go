package models

import (
	"time"
)

type SubscriptionStatus string

const (
	StatusRequested SubscriptionStatus = "requested"
	StatusApproved  SubscriptionStatus = "approved"
	StatusActive    SubscriptionStatus = "active"
	StatusInactive  SubscriptionStatus = "inactive"
	StatusExpired   SubscriptionStatus = "expired"
)

type Subscription struct {
	ID            uint               `json:"id" gorm:"primaryKey"`
	CustomerID    uint               `json:"customer_id" gorm:"not null"`
	PackID        uint               `json:"pack_id" gorm:"not null"`
	Status        SubscriptionStatus `json:"status" gorm:"default:'requested'"`
	RequestedAt   time.Time          `json:"requested_at"`
	ApprovedAt    *time.Time         `json:"approved_at"`
	AssignedAt    *time.Time         `json:"assigned_at"`
	ExpiresAt     *time.Time         `json:"expires_at"`
	DeactivatedAt *time.Time         `json:"deactivated_at"`
	CreatedAt     time.Time          `json:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at"`
	
	// Relationships
	Customer *Customer         `json:"customer,omitempty" gorm:"foreignKey:CustomerID"`
	Pack     *SubscriptionPack `json:"pack,omitempty" gorm:"foreignKey:PackID"`
}

// CanTransitionTo checks if the subscription can transition to the given status
func (s *Subscription) CanTransitionTo(newStatus SubscriptionStatus) bool {
	validTransitions := map[SubscriptionStatus][]SubscriptionStatus{
		StatusRequested: {StatusApproved, StatusInactive},
		StatusApproved:  {StatusActive, StatusInactive},
		StatusActive:    {StatusInactive, StatusExpired},
		StatusInactive:  {StatusActive},
		StatusExpired:   {StatusRequested},
	}
	
	allowedStatuses, exists := validTransitions[s.Status]
	if !exists {
		return false
	}
	
	for _, allowed := range allowedStatuses {
		if allowed == newStatus {
			return true
		}
	}
	return false
}

// IsActive checks if the subscription is currently active
func (s *Subscription) IsActive() bool {
	return s.Status == StatusActive && s.ExpiresAt != nil && s.ExpiresAt.After(time.Now())
}

// IsExpired checks if the subscription has expired
func (s *Subscription) IsExpired() bool {
	return s.ExpiresAt != nil && s.ExpiresAt.Before(time.Now())
}

// CalculateExpiry calculates the expiry date based on pack validity
func (s *Subscription) CalculateExpiry(pack *SubscriptionPack) {
	if s.AssignedAt != nil {
		expiry := s.AssignedAt.AddDate(0, pack.ValidityMonths, 0)
		s.ExpiresAt = &expiry
	}
}
