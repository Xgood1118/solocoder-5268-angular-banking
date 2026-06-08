package recon

import (
	"banking/internal/account"
	"banking/pkg/cache"
	"context"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

type Service struct {
	repo        *Repository
	accountRepo *account.Repository
	cron        *cron.Cron
	cache       cache.Cache
}

func NewService(repo *Repository, accountRepo *account.Repository, c cache.Cache) *Service {
	return &Service{
		repo:        repo,
		accountRepo: accountRepo,
		cron:        cron.New(),
		cache:       c,
	}
}

func (s *Service) StartScheduler() {
	s.cron.AddFunc("0 0 0 * * *", func() {
		s.runDailyRecon()
	})
	s.cron.Start()
	log.Println("Recon scheduler started")
}

func (s *Service) runDailyRecon() {
	lockKey := "recon:lock:daily"
	ctx := context.Background()

	ok, err := s.cache.SetNX(ctx, lockKey, "1", 30*time.Minute)
	if err != nil || !ok {
		log.Println("Recon already running, skip")
		return
	}
	defer s.cache.Del(ctx, lockKey)

	log.Println("Starting daily reconciliation...")

	reconDate := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	var accounts []account.Account
	s.accountRepo.DB().Where("status != ?", "closed").Find(&accounts)

	successCount := 0
	diffCount := 0

	for _, acc := range accounts {
		report, err := s.reconcileAccount(acc.ID, reconDate)
		if err != nil {
			log.Printf("Recon failed for account %d: %v", acc.ID, err)
			continue
		}
		if report.Status == ReconStatusSuccess {
			successCount++
		} else {
			diffCount++
		}
	}

	log.Printf("Daily recon completed: %d success, %d differences", successCount, diffCount)
}

func (s *Service) reconcileAccount(accountID uint, reconDate string) (*ReconReport, error) {
	db := s.repo.db

	var acc account.Account
	if err := db.First(&acc, accountID).Error; err != nil {
		return nil, err
	}

	var totalCredit float64
	var totalDebit float64

	db.Model(&account.LedgerEntry{}).
		Where("account_id = ? AND DATE(created_at) <= ? AND entry_type = ?",
			accountID, reconDate, "credit").
		Select("COALESCE(SUM(amount), 0)").
		Scan(&totalCredit)

	db.Model(&account.LedgerEntry{}).
		Where("account_id = ? AND DATE(created_at) <= ? AND entry_type = ?",
			accountID, reconDate, "debit").
		Select("COALESCE(SUM(amount), 0)").
		Scan(&totalDebit)

	ledgerBalance := totalCredit - totalDebit

	systemBalance := acc.Balance

	diff := math.Abs(systemBalance - ledgerBalance)

	var entryCount int64
	db.Model(&account.LedgerEntry{}).
		Where("account_id = ? AND DATE(created_at) <= ?", accountID, reconDate).
		Count(&entryCount)

	report := &ReconReport{
		ReconDate:     reconDate,
		AccountID:     accountID,
		AccountNo:     acc.AccountNumber,
		SystemBalance: systemBalance,
		LedgerBalance: ledgerBalance,
		Difference:    diff,
		TotalEntries:  int(entryCount),
	}

	if diff < 0.01 {
		report.Status = ReconStatusSuccess
	} else {
		report.Status = ReconStatusDifference
		s.generateDifferences(accountID, reconDate, report)
	}

	if err := db.Create(report).Error; err != nil {
		return nil, err
	}

	return report, nil
}

func (s *Service) generateDifferences(accountID uint, reconDate string, report *ReconReport) {
	db := s.repo.db

	var entries []account.LedgerEntry
	db.Where("account_id = ? AND DATE(created_at) = ?", accountID, reconDate).
		Order("created_at").Find(&entries)

	var runningBalance float64
	var acc account.Account
	db.First(&acc, accountID)

	var prevBalance float64
	db.Model(&account.LedgerEntry{}).
		Where("account_id = ? AND DATE(created_at) < ?", accountID, reconDate).
		Select("COALESCE(SUM(CASE WHEN entry_type = 'credit' THEN amount ELSE -amount END), 0)").
		Scan(&prevBalance)

	runningBalance = prevBalance

	for _, entry := range entries {
		if entry.EntryType == "credit" {
			runningBalance += entry.Amount
		} else {
			runningBalance -= entry.Amount
		}

		if math.Abs(runningBalance-entry.BalanceAfter) > 0.01 {
			diff := &ReconDifference{
				AccountID:      accountID,
				DiffType:       DiffTypeAmountMismatch,
				TransactionID:  entry.TransactionID,
				ExpectedAmount: runningBalance,
				ActualAmount:   entry.BalanceAfter,
				Description:    fmt.Sprintf("交易%s余额不符，预期%.2f，实际%.2f", entry.TransactionID, runningBalance, entry.BalanceAfter),
			}
			db.Create(diff)
		}
	}
}

func (s *Service) ManualReconcile(diffID uint, operatorID uint) error {
	db := s.repo.db

	var diff ReconDifference
	if err := db.First(&diff, diffID).Error; err != nil {
		return err
	}

	if diff.IsManualRecon {
		return fmt.Errorf("already reconciled")
	}

	now := time.Now()
	diff.IsManualRecon = true
	diff.ManualReconBy = operatorID
	diff.ManualReconAt = &now

	return db.Save(&diff).Error
}

func (s *Service) GetReports(accountID uint, page, pageSize int) (int64, []ReconReport, error) {
	var total int64
	var reports []ReconReport

	db := s.repo.db.Model(&ReconReport{})
	if accountID > 0 {
		db = db.Where("account_id = ?", accountID)
	}

	db.Count(&total)

	offset := (page - 1) * pageSize
	err := db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&reports).Error

	return total, reports, err
}

func (s *Service) GetDifferences(reportID uint, page, pageSize int) (int64, []ReconDifference, error) {
	var total int64
	var diffs []ReconDifference

	db := s.repo.db.Model(&ReconDifference{}).Where("report_id = ?", reportID)

	db.Count(&total)

	offset := (page - 1) * pageSize
	err := db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&diffs).Error

	return total, diffs, err
}

func (s *Service) TriggerRecon(accountID uint) (*ReconReport, error) {
	reconDate := time.Now().Format("2006-01-02")
	return s.reconcileAccount(accountID, reconDate)
}
