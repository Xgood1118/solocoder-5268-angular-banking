package auth

import (
	"time"
)

type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Username     string    `gorm:"uniqueIndex;size:50;not null" json:"username"`
	PasswordHash string    `gorm:"size:255;not null" json:"-"`
	Email        string    `gorm:"size:100" json:"email"`
	Phone        string    `gorm:"size:20" json:"phone"`
	FullName     string    `gorm:"size:100" json:"full_name"`
	IDCard       string    `gorm:"size:20" json:"id_card"`
	Status       string    `gorm:"size:20;default:active" json:"status"`
	TwoFAEnabled bool      `gorm:"default:false" json:"twofa_enabled"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type VerificationCode struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"index"`
	Type      string    `gorm:"size:20"`
	Code      string    `gorm:"size:10"`
	Target    string    `gorm:"size:100"`
	Attempts  int       `gorm:"default:0"`
	ExpiresAt time.Time `gorm:"index"`
	CreatedAt time.Time
	Locked    bool `gorm:"default:false"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=8"`
	Email    string `json:"email" binding:"required,email"`
	Phone    string `json:"phone" binding:"required"`
	FullName string `json:"full_name" binding:"required"`
	IDCard   string `json:"id_card" binding:"required"`
}

type TwoFARequest struct {
	Code string `json:"code" binding:"required,len=6"`
}

type SendCodeRequest struct {
	Type   string `json:"type" binding:"required,oneof=sms email"`
	Target string `json:"target"`
}

type LoginResponse struct {
	Token      string `json:"token"`
	User       *User  `json:"user"`
	NeedTwoFA  bool   `json:"need_twofa"`
	TwoFAToken string `json:"twofa_token,omitempty"`
}
