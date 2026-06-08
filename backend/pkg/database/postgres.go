package database

import (
	"banking/internal/account"
	"banking/internal/audit"
	"banking/internal/auth"
	"banking/internal/limit"
	"banking/internal/recon"
	"banking/internal/transfer"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewPostgres(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&auth.User{},
		&account.Account{},
		&account.LedgerEntry{},
		&transfer.Transfer{},
		&audit.AuditLog{},
		&limit.LimitConfig{},
		&limit.LimitUsage{},
		&auth.VerificationCode{},
		&recon.ReconReport{},
		&recon.ReconDifference{},
	)
}
