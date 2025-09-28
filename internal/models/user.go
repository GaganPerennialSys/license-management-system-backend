package models

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Email        string    `json:"email" gorm:"uniqueIndex;not null"`
	Password     string    `json:"-" gorm:"not null"`
	Role         string    `json:"role" gorm:"default:'customer'"`
	APIKey       *string   `json:"-" gorm:"uniqueIndex"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	
	// Relationships
	Customer *Customer `json:"customer,omitempty" gorm:"foreignKey:UserID"`
}

func (u *User) HashPassword() error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

func (u *User) IsAdmin() bool {
	return u.Role == "admin"
}

func (u *User) IsCustomer() bool {
	return u.Role == "customer"
}

// GenerateAPIKey generates a new API key for the user
func (u *User) GenerateAPIKey() error {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return err
	}
	apiKey := "sk-sdk-" + hex.EncodeToString(bytes)
	u.APIKey = &apiKey
	return nil
}

// HasAPIKey checks if the user has an API key
func (u *User) HasAPIKey() bool {
	return u.APIKey != nil && *u.APIKey != ""
}
