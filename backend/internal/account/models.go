package account

import (
	"time"
)

type AccountType string

const (
	AccountTypeSavings    AccountType = "savings"
	AccountTypeFixedDeposit AccountType = "fixed_deposit"
	AccountTypeChecking   AccountType = "checking"
)

type AccountStatus string

const (
	AccountStatusActive   AccountStatus = "active"
	AccountStatusFrozen   AccountStatus = "frozen"
	AccountStatusClosed   AccountStatus = "closed"
)

type Currency string

const (
	CurrencyCNY Currency = "CNY"
	CurrencyUSD Currency = "USD"
	CurrencyEUR Currency = "EUR"
)

type Account struct {
	ID              uint          `gorm:"primaryKey" json:"id"`
	UserID          uint          `gorm:"index;not null" json:"user_id"`
	AccountNumber   string        `gorm:"uniqueIndex;size:20;not null" json:"account_number"`
	AccountType     AccountType   `gorm:"size:20;not null" json:"account_type"`
	AccountName     string        `gorm:"size:100" json:"account_name"`
	Currency        Currency      `gorm:"size:5;default:CNY" json:"currency"`
	Status          AccountStatus `gorm:"size:20;default:active" json:"status"`
	Balance         float64       `gorm:"type:decimal(18,2);default:0" json:"balance"`
	AvailableBalance float64      `gorm:"type:decimal(18,2);default:0" json:"available_balance"`
	FrozenAmount    float64       `gorm:"type:decimal(18,2);default:0" json:"frozen_amount"`
	InterestRate    float64       `gorm:"type:decimal(5,4)" json:"interest_rate"`
	TermDays        int           `gorm:"default:0" json:"term_days"`
	MaturityDate    *time.Time    `json:"maturity_date,omitempty"`
	Version         int           `gorm:"default:0" json:"version"`
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`
	ClosedAt        *time.Time    `json:"closed_at,omitempty"`
}

type LedgerEntry struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	AccountID     uint      `gorm:"index;not null" json:"account_id"`
	TransactionID string    `gorm:"size:36;index" json:"transaction_id"`
	BizID         string    `gorm:"size:36;uniqueIndex" json:"biz_id"`
	EntryType     string    `gorm:"size:10;not null" json:"entry_type"`
	Amount        float64   `gorm:"type:decimal(18,2);not null" json:"amount"`
	BalanceAfter  float64   `gorm:"type:decimal(18,2)" json:"balance_after"`
	Description   string    `gorm:"size:255" json:"description"`
	RefAccountID  uint      `gorm:"index" json:"ref_account_id"`
	CreatedAt     time.Time `json:"created_at"`
}

type OpenAccountRequest struct {
	UserID      uint        `json:"-"`
	AccountType AccountType `json:"account_type" binding:"required,oneof=savings fixed_deposit checking"`
	AccountName string      `json:"account_name" binding:"required"`
	Currency    Currency    `json:"currency"`
	Amount      float64     `json:"amount"`
	TermDays    int         `json:"term_days"`
}

type TransferRequest struct {
	FromAccountID uint    `json:"from_account_id" binding:"required"`
	ToAccountID   uint    `json:"to_account_id" binding:"required"`
	Amount        float64 `json:"amount" binding:"required,gt=0"`
	Description   string  `json:"description"`
	BizID         string  `json:"biz_id" binding:"required"`
}

type FreezeRequest struct {
	AccountID uint    `json:"account_id" binding:"required"`
	Amount    float64 `json:"amount" binding:"required,gt=0"`
	Reason    string  `json:"reason"`
}
