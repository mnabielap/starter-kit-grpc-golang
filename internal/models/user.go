package models

import (
	"time"

	"starter-kit-grpc-golang/pkg/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID              string    `gorm:"type:uuid;primary_key;"` // Stored as string to match Proto
	Name            string    `gorm:"not null"`
	Email           string    `gorm:"uniqueIndex;not null"`
	Password        string    `gorm:"not null"`
	Role            string    `gorm:"default:'user'"`
	IsEmailVerified bool      `gorm:"default:false"`
	CreatedAt       time.Time `gorm:"autoCreateTime"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime"`
}

// BeforeCreate generates a UUID if one doesn't exist
func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == "" {
		u.ID = uuid.New().String()
	}
	return
}

// BeforeSave hashes the password if it's not already hashed (simple length check)
// In a real production app, checking Changed() is more robust, but this matches the starter kit.
func (u *User) BeforeSave(tx *gorm.DB) (err error) {
	if u.Password != "" && len(u.Password) < 60 {
		hashed, err := utils.HashPassword(u.Password)
		if err != nil {
			return err
		}
		u.Password = hashed
	}
	return
}

// ComparePassword is a helper to verify login
func (u *User) ComparePassword(plainPassword string) bool {
	return utils.CheckPassword(plainPassword, u.Password)
}