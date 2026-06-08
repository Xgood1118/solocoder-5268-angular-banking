package report

import (
	"banking/internal/account"
	"banking/internal/transfer"
	"time"

	"gorm.io/gorm"
)

type ReportService struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *ReportService {
	return &ReportService{db: db}
}

type BalanceReport struct {
	Date            string  `json:"date"`
	TotalBalance    float64 `json:"total_balance"`
	AccountCount    int     `json:"account_count"`
	TotalDeposit    float64 `json:"total_deposit"`
	TotalWithdrawal float64 `json:"total_withdrawal"`
	NetChange       float64 `json:"net_change"`
}

type TransactionReport struct {
	Date         string  `json:"date"`
	TotalCount   int     `json:"total_count"`
	TotalAmount  float64 `json:"total_amount"`
	SuccessCount int     `json:"success_count"`
	FailedCount  int     `json:"failed_count"`
	FeeIncome    float64 `json:"fee_income"`
}

func (s *ReportService) GetDailyBalanceReport(userID uint, date string) (*BalanceReport, error) {
	var totalBalance float64
	var accountCount int64

	s.db.Model(&account.Account{}).
		Where("user_id = ? AND status != ?", userID, "closed").
		Select("COALESCE(SUM(balance), 0)").Scan(&totalBalance)

	s.db.Model(&account.Account{}).
		Where("user_id = ? AND status != ?", userID, "closed").
		Count(&accountCount)

	startDate, _ := time.Parse("2006-01-02", date)
	endDate := startDate.AddDate(0, 0, 1)

	var totalDeposit float64
	s.db.Model(&account.LedgerEntry{}).
		Joins("JOIN accounts ON accounts.id = ledger_entries.account_id").
		Where("accounts.user_id = ? AND ledger_entries.entry_type = ? AND ledger_entries.created_at >= ? AND ledger_entries.created_at < ?",
			userID, "credit", startDate, endDate).
		Select("COALESCE(SUM(ledger_entries.amount), 0)").Scan(&totalDeposit)

	var totalWithdrawal float64
	s.db.Model(&account.LedgerEntry{}).
		Joins("JOIN accounts ON accounts.id = ledger_entries.account_id").
		Where("accounts.user_id = ? AND ledger_entries.entry_type = ? AND ledger_entries.created_at >= ? AND ledger_entries.created_at < ?",
			userID, "debit", startDate, endDate).
		Select("COALESCE(SUM(ledger_entries.amount), 0)").Scan(&totalWithdrawal)

	return &BalanceReport{
		Date:            date,
		TotalBalance:    totalBalance,
		AccountCount:    int(accountCount),
		TotalDeposit:    totalDeposit,
		TotalWithdrawal: totalWithdrawal,
		NetChange:       totalDeposit - totalWithdrawal,
	}, nil
}

func (s *ReportService) GetWeeklyBalanceReport(userID uint) ([]BalanceReport, error) {
	var reports []BalanceReport
	today := time.Now()

	for i := 6; i >= 0; i-- {
		date := today.AddDate(0, 0, -i).Format("2006-01-02")
		report, _ := s.GetDailyBalanceReport(userID, date)
		reports = append(reports, *report)
	}

	return reports, nil
}

func (s *ReportService) GetTransactionReport(userID uint, startDate, endDate string) (*TransactionReport, error) {
	start, _ := time.Parse("2006-01-02", startDate)
	end, _ := time.Parse("2006-01-02", endDate)
	end = end.AddDate(0, 0, 1)

	var totalCount int64
	var totalAmount float64
	var successCount int64
	var failedCount int64
	var feeIncome float64

	s.db.Model(&transfer.Transfer{}).
		Where("user_id = ? AND created_at >= ? AND created_at < ?", userID, start, end).
		Count(&totalCount)

	s.db.Model(&transfer.Transfer{}).
		Where("user_id = ? AND status = ? AND created_at >= ? AND created_at < ?", userID, "success", start, end).
		Select("COALESCE(SUM(amount), 0), COALESCE(SUM(fee), 0), COUNT(*)").
		Row().Scan(&totalAmount, &feeIncome, &successCount)

	failedCount = totalCount - successCount

	return &TransactionReport{
		Date:         startDate + " ~ " + endDate,
		TotalCount:   int(totalCount),
		TotalAmount:  totalAmount,
		SuccessCount: int(successCount),
		FailedCount:  int(failedCount),
		FeeIncome:    feeIncome,
	}, nil
}
