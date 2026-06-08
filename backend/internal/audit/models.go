package audit

import (
	"time"
)

type AuditLog struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"index" json:"user_id"`
	Action      string    `gorm:"size:50;index" json:"action"`
	Module      string    `gorm:"size:30;index" json:"module"`
	Description string    `gorm:"type:text" json:"description"`
	IPAddress   string    `gorm:"size:50" json:"ip_address"`
	UserAgent   string    `gorm:"size:255" json:"user_agent"`
	HMAC        string    `gorm:"size:128" json:"-"`
	PrevLogHMAC string    `gorm:"size:128" json:"-"`
	CreatedAt   time.Time `json:"created_at"`
}

type AuditQuery struct {
	UserID    uint
	Action    string
	Module    string
	StartTime time.Time
	EndTime   time.Time
	Page      int
	PageSize  int
}

type AuditLogResponse struct {
	Total    int64      `json:"total"`
	Page     int        `json:"page"`
	PageSize int        `json:"page_size"`
	Logs     []AuditLog `json:"logs"`
}
