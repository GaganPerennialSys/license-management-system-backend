package models

import (
	"time"

	"gorm.io/gorm"
)

type Customer struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UserID    uint           `json:"user_id" gorm:"uniqueIndex;not null"`
	Name      string         `json:"name" gorm:"not null"`
	Phone     string         `json:"phone"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
	
	// Relationships
	User          *User           `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Subscriptions []*Subscription `json:"subscriptions,omitempty" gorm:"foreignKey:CustomerID"`
}

// GetActiveSubscription returns the currently active subscription for this customer
func (c *Customer) GetActiveSubscription(db *gorm.DB) (*Subscription, error) {
	var subscription Subscription
	err := db.Where("customer_id = ? AND status = ?", c.ID, "active").First(&subscription).Error
	if err != nil {
		return nil, err
	}
	return &subscription, nil
}

// HasActiveSubscription checks if customer has an active subscription
func (c *Customer) HasActiveSubscription(db *gorm.DB) bool {
	var count int64
	db.Model(&Subscription{}).Where("customer_id = ? AND status = ?", c.ID, "active").Count(&count)
	return count > 0
}
