package transfer

import (
	"time"

	"gorm.io/gorm"
)

type TransferType string

const (
	TransferTypeIntraBank TransferType = "intra_bank"
	TransferTypeInterBank TransferType = "inter_bank"
)

type TransferSpeed string

const (
	TransferSpeedRealTime TransferSpeed = "realtime"
	TransferSpeedNormal   TransferSpeed = "normal"
)

type TransferStatus string

const (
	StatusPending   TransferStatus = "pending"
	StatusFrozen    TransferStatus = "frozen"
	StatusProcessing TransferStatus = "processing"
	StatusSuccess   TransferStatus = "success"
	StatusFailed    TransferStatus = "failed"
	StatusRolledBack TransferStatus = "rolled_back"
)

type Transfer struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	BizID           string         `gorm:"uniqueIndex;size:36;not null" json:"biz_id"`
	UserID          uint           `gorm:"index" json:"user_id"`
	FromAccountID   uint           `gorm:"index;not null" json:"from_account_id"`
	ToAccountID     uint           `gorm:"index" json:"to_account_id"`
	FromAccountNo   string         `gorm:"size:30" json:"from_account_no"`
	ToAccountNo     string         `gorm:"size:30" json:"to_account_no"`
	ToBankName      string         `gorm:"size:100" json:"to_bank_name"`
	ToAccountName   string         `gorm:"size:100" json:"to_account_name"`
	Amount          float64        `gorm:"type:decimal(18,2);not null" json:"amount"`
	Currency        string         `gorm:"size:5;default:CNY" json:"currency"`
	TransferType    TransferType   `gorm:"size:20" json:"transfer_type"`
	TransferSpeed   TransferSpeed  `gorm:"size:20" json:"transfer_speed"`
	Status          TransferStatus `gorm:"size:20;index" json:"status"`
	Description     string         `gorm:"size:255" json:"description"`
	Fee             float64        `gorm:"type:decimal(10,2);default:0" json:"fee"`
	ClearingRefNo   string         `gorm:"size:50" json:"clearing_ref_no"`
	FailureReason   string         `gorm:"size:255" json:"failure_reason"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	CompletedAt     *time.Time     `json:"completed_at,omitempty"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

type CreateTransferRequest struct {
	UserID        uint          `json:"-"`
	FromAccountID uint          `json:"from_account_id" binding:"required"`
	ToAccountID   uint          `json:"to_account_id"`
	ToAccountNo   string        `json:"to_account_no"`
	ToBankName    string        `json:"to_bank_name"`
	ToAccountName string        `json:"to_account_name" binding:"required"`
	Amount        float64       `json:"amount" binding:"required,gt=0"`
	TransferType  TransferType  `json:"transfer_type" binding:"required,oneof=intra_bank inter_bank"`
	TransferSpeed TransferSpeed `json:"transfer_speed" binding:"required,oneof=realtime normal"`
	Description   string        `json:"description"`
	BizID         string        `json:"biz_id" binding:"required"`
}
