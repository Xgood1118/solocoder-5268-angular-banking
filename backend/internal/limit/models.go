package limit

import (
	"time"
)

type LimitType string

const (
	LimitTypePerTransaction LimitType = "per_transaction"
	LimitTypeDaily          LimitType = "daily"
	LimitTypeMonthly        LimitType = "monthly"
)

type LimitConfig struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"index;default:0" json:"user_id"`
	LimitType   LimitType `gorm:"size:20;index" json:"limit_type"`
	Amount      float64   `gorm:"type:decimal(18,2)" json:"amount"`
	Currency    string    `gorm:"size:5;default:CNY" json:"currency"`
	Scope       string    `gorm:"size:50;default:transfer" json:"scope"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type LimitUsage struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index" json:"user_id"`
	LimitType LimitType `gorm:"size:20" json:"limit_type"`
	Amount    float64   `gorm:"type:decimal(18,2)" json:"amount"`
	Scope     string    `gorm:"size:50" json:"scope"`
	PeriodKey string    `gorm:"size:20;index" json:"period_key"`
	CreatedAt time.Time `json:"created_at"`
}
