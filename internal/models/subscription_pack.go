package models

import (
	"time"

	"gorm.io/gorm"
)

type SubscriptionPack struct {
	ID             uint           `json:"id" gorm:"primaryKey"`
	Name           string         `json:"name" gorm:"not null"`
	Description    string         `json:"description"`
	SKU            string         `json:"sku" gorm:"uniqueIndex;not null"`
	Price          float64        `json:"price" gorm:"type:decimal(10,2);not null"`
	ValidityMonths int            `json:"validity_months" gorm:"not null;check:validity_months >= 1 AND validity_months <= 12"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"deleted_at" gorm:"index"`
	
	// Relationships
	Subscriptions []*Subscription `json:"subscriptions,omitempty" gorm:"foreignKey:PackID"`
}

// IsValid checks if the subscription pack is valid (not deleted)
func (sp *SubscriptionPack) IsValid() bool {
	return sp.DeletedAt.Time.IsZero()
}
