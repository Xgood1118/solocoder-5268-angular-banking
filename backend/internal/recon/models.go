package recon

import (
	"time"
)

type ReconStatus string

const (
	ReconStatusSuccess    ReconStatus = "success"
	ReconStatusDifference ReconStatus = "difference"
	ReconStatusFailed     ReconStatus = "failed"
)

type DiffType string

const (
	DiffTypeAmountMismatch DiffType = "amount_mismatch"
	DiffTypeMissingEntry   DiffType = "missing_entry"
	DiffTypeExtraEntry     DiffType = "extra_entry"
)

type ReconReport struct {
	ID           uint         `gorm:"primaryKey" json:"id"`
	ReconDate    string       `gorm:"size:10;index" json:"recon_date"`
	AccountID    uint         `gorm:"index" json:"account_id"`
	AccountNo    string       `gorm:"size:30" json:"account_no"`
	SystemBalance float64     `gorm:"type:decimal(18,2)" json:"system_balance"`
	LedgerBalance float64     `gorm:"type:decimal(18,2)" json:"ledger_balance"`
	Difference   float64      `gorm:"type:decimal(18,2)" json:"difference"`
	Status       ReconStatus  `gorm:"size:20" json:"status"`
	TotalEntries int          `json:"total_entries"`
	CreatedAt    time.Time    `json:"created_at"`
}

type ReconDifference struct {
	ID             uint     `gorm:"primaryKey" json:"id"`
	ReportID       uint     `gorm:"index" json:"report_id"`
	AccountID      uint     `gorm:"index" json:"account_id"`
	DiffType       DiffType `gorm:"size:30" json:"diff_type"`
	TransactionID  string   `gorm:"size:50" json:"transaction_id"`
	ExpectedAmount float64  `gorm:"type:decimal(18,2)" json:"expected_amount"`
	ActualAmount   float64  `gorm:"type:decimal(18,2)" json:"actual_amount"`
	Description    string   `gorm:"size:255" json:"description"`
	IsManualRecon  bool     `gorm:"default:false" json:"is_manual_recon"`
	ManualReconBy  uint     `json:"manual_recon_by"`
	ManualReconAt  *time.Time `json:"manual_recon_at,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}
