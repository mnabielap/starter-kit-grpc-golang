package models

import (
	"time"
)

const (
	TokenTypeRefresh       = "refresh"
	TokenTypeResetPassword = "resetPassword"
	TokenTypeVerifyEmail   = "verifyEmail"
)

type Token struct {
	ID          uint      `gorm:"primary_key"`
	Token       string    `gorm:"index;not null"`
	UserID      string    `gorm:"type:uuid;not null"`
	User        User      `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Type        string    `gorm:"not null"`
	Expires     time.Time `gorm:"not null"`
	Blacklisted bool      `gorm:"default:false"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
}