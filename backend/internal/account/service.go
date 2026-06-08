package account

import (
	"banking/internal/audit"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

type Service struct {
	repo     *Repository
	auditSvc *audit.Service
}

func NewService(repo *Repository, auditSvc *audit.Service) *Service {
	return &Service{repo: repo, auditSvc: auditSvc}
}

func (s *Service) OpenAccount(req *OpenAccountRequest) (*Account, error) {
	accountNumber := s.generateAccountNumber(req.AccountType)

	interestRate := 0.0
	switch req.AccountType {
	case AccountTypeSavings:
		interestRate = 0.025
	case AccountTypeFixedDeposit:
		if req.TermDays >= 365 {
			interestRate = 0.035
		} else if req.TermDays >= 180 {
			interestRate = 0.030
		} else {
			interestRate = 0.028
		}
	case AccountTypeChecking:
		interestRate = 0.005
	}

	var maturityDate *time.Time
	if req.AccountType == AccountTypeFixedDeposit && req.TermDays > 0 {
		t := time.Now().AddDate(0, 0, req.TermDays)
		maturityDate = &t
	}

	account := &Account{
		UserID:          req.UserID,
		AccountNumber:   accountNumber,
		AccountType:     req.AccountType,
		AccountName:     req.AccountName,
		Currency:        req.Currency,
		Status:          AccountStatusActive,
		Balance:         0,
		AvailableBalance: 0,
		FrozenAmount:    0,
		InterestRate:    interestRate,
		TermDays:        req.TermDays,
		MaturityDate:    maturityDate,
	}

	tx := s.repo.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	if err := tx.Create(account).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if req.Amount > 0 {
		bizID := uuid.New().String()
		ledger := &LedgerEntry{
			AccountID:    account.ID,
			BizID:        bizID,
			EntryType:    "credit",
			Amount:       req.Amount,
			BalanceAfter: req.Amount,
			Description:  "开户存入",
		}
		if err := tx.Create(ledger).Error; err != nil {
			tx.Rollback()
			return nil, err
		}

		account.Balance = req.Amount
		account.AvailableBalance = req.Amount
		if err := tx.Save(account).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	s.auditSvc.Log(req.UserID, "open_account", "account", fmt.Sprintf("opened account %s", accountNumber))

	return account, nil
}

func (s *Service) CloseAccount(userID, accountID uint) error {
	var account Account
	if err := s.repo.db.First(&account, accountID).Error; err != nil {
		return err
	}

	if account.UserID != userID {
		return errors.New("permission denied")
	}

	if account.Status != AccountStatusActive {
		return errors.New("account is not active")
	}

	if account.Balance > 0 {
		return errors.New("account has balance, please withdraw first")
	}

	now := time.Now()
	account.Status = AccountStatusClosed
	account.ClosedAt = &now

	if err := s.repo.db.Save(&account).Error; err != nil {
		return err
	}

	s.auditSvc.Log(userID, "close_account", "account", fmt.Sprintf("closed account %s", account.AccountNumber))

	return nil
}

func (s *Service) FreezeAccount(userID, accountID uint, reason string) error {
	var account Account
	if err := s.repo.db.First(&account, accountID).Error; err != nil {
		return err
	}

	if account.UserID != userID {
		return errors.New("permission denied")
	}

	if account.Status != AccountStatusActive {
		return errors.New("account is not active")
	}

	account.Status = AccountStatusFrozen
	if err := s.repo.db.Save(&account).Error; err != nil {
		return err
	}

	s.auditSvc.Log(userID, "freeze_account", "account", fmt.Sprintf("frozen account %s, reason: %s", account.AccountNumber, reason))

	return nil
}

func (s *Service) UnfreezeAccount(userID, accountID uint) error {
	var account Account
	if err := s.repo.db.First(&account, accountID).Error; err != nil {
		return err
	}

	if account.UserID != userID {
		return errors.New("permission denied")
	}

	if account.Status != AccountStatusFrozen {
		return errors.New("account is not frozen")
	}

	account.Status = AccountStatusActive
	if err := s.repo.db.Save(&account).Error; err != nil {
		return err
	}

	s.auditSvc.Log(userID, "unfreeze_account", "account", fmt.Sprintf("unfrozen account %s", account.AccountNumber))

	return nil
}

func (s *Service) GetAccount(userID, accountID uint) (*Account, error) {
	var account Account
	if err := s.repo.db.First(&account, accountID).Error; err != nil {
		return nil, err
	}

	if account.UserID != userID {
		return nil, errors.New("permission denied")
	}

	return &account, nil
}

func (s *Service) GetAccountByID(accountID uint) (*Account, error) {
	var account Account
	if err := s.repo.db.First(&account, accountID).Error; err != nil {
		return nil, err
	}
	return &account, nil
}

func (s *Service) GetAccountByNumber(accountNumber string) (*Account, error) {
	var account Account
	if err := s.repo.db.Where("account_number = ?", accountNumber).First(&account).Error; err != nil {
		return nil, err
	}
	return &account, nil
}

func (s *Service) ListAccounts(userID uint) ([]Account, error) {
	var accounts []Account
	if err := s.repo.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&accounts).Error; err != nil {
		return nil, err
	}
	return accounts, nil
}

func (s *Service) Debit(tx *gorm.DB, accountID uint, amount float64, bizID, description, transactionID string) error {
	var account Account
	if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&account, accountID).Error; err != nil {
		return err
	}

	if account.Status != AccountStatusActive {
		return errors.New("account is not active")
	}

	if account.AvailableBalance < amount {
		return errors.New("insufficient balance")
	}

	entry := &LedgerEntry{
		AccountID:     accountID,
		BizID:         bizID,
		TransactionID: transactionID,
		EntryType:     "debit",
		Amount:        amount,
		Description:   description,
		BalanceAfter:  account.Balance - amount,
	}

	if err := tx.Create(entry).Error; err != nil {
		return err
	}

	result := tx.Model(&Account{}).
		Where("id = ? AND version = ?", accountID, account.Version).
		Updates(map[string]interface{}{
			"balance":           gorm.Expr("balance - ?", amount),
			"available_balance": gorm.Expr("available_balance - ?", amount),
			"version":           account.Version + 1,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("concurrent update conflict, please retry")
	}

	return nil
}

func (s *Service) Credit(tx *gorm.DB, accountID uint, amount float64, bizID, description, transactionID string, refAccountID uint) error {
	var account Account
	if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&account, accountID).Error; err != nil {
		return err
	}

	if account.Status == AccountStatusClosed {
		return errors.New("account is closed")
	}

	entry := &LedgerEntry{
		AccountID:     accountID,
		BizID:         bizID,
		TransactionID: transactionID,
		EntryType:     "credit",
		Amount:        amount,
		Description:   description,
		BalanceAfter:  account.Balance + amount,
		RefAccountID:  refAccountID,
	}

	if err := tx.Create(entry).Error; err != nil {
		return err
	}

	result := tx.Model(&Account{}).
		Where("id = ? AND version = ?", accountID, account.Version).
		Updates(map[string]interface{}{
			"balance":           gorm.Expr("balance + ?", amount),
			"available_balance": gorm.Expr("available_balance + ?", amount),
			"version":           account.Version + 1,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("concurrent update conflict, please retry")
	}

	return nil
}

func (s *Service) FreezeAmount(accountID uint, amount float64, bizID string) error {
	tx := s.repo.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	if err := s.FreezeAmountTx(tx, accountID, amount); err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}

func (s *Service) FreezeAmountTx(tx *gorm.DB, accountID uint, amount float64) error {
	var account Account
	if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&account, accountID).Error; err != nil {
		return err
	}

	if account.AvailableBalance < amount {
		return errors.New("insufficient available balance")
	}

	result := tx.Model(&account).
		Where("id = ? AND version = ?", accountID, account.Version).
		Updates(map[string]interface{}{
			"frozen_amount":     gorm.Expr("frozen_amount + ?", amount),
			"available_balance": gorm.Expr("available_balance - ?", amount),
			"version":           account.Version + 1,
		})

	if result.Error != nil || result.RowsAffected == 0 {
		return errors.New("freeze failed")
	}

	return nil
}

func (s *Service) UnfreezeAmount(accountID uint, amount float64) error {
	tx := s.repo.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	if err := s.UnfreezeAmountTx(tx, accountID, amount); err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}

func (s *Service) UnfreezeAmountTx(tx *gorm.DB, accountID uint, amount float64) error {
	var account Account
	if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&account, accountID).Error; err != nil {
		return err
	}

	if account.FrozenAmount < amount {
		return errors.New("frozen amount insufficient")
	}

	result := tx.Model(&account).
		Where("id = ? AND version = ?", accountID, account.Version).
		Updates(map[string]interface{}{
			"frozen_amount":     gorm.Expr("frozen_amount - ?", amount),
			"available_balance": gorm.Expr("available_balance + ?", amount),
			"version":           account.Version + 1,
		})

	if result.Error != nil || result.RowsAffected == 0 {
		return errors.New("unfreeze failed")
	}

	return nil
}

func (s *Service) GetLedger(accountID uint, page, pageSize int) (int64, []LedgerEntry, error) {
	var total int64
	var entries []LedgerEntry

	s.repo.db.Model(&LedgerEntry{}).Where("account_id = ?", accountID).Count(&total)

	offset := (page - 1) * pageSize
	err := s.repo.db.Where("account_id = ?", accountID).
		Order("created_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&entries).Error

	return total, entries, err
}

func (s *Service) GetBalance(accountID uint) (float64, float64, error) {
	var account Account
	if err := s.repo.db.Select("balance, available_balance").First(&account, accountID).Error; err != nil {
		return 0, 0, err
	}
	return account.Balance, account.AvailableBalance, nil
}

func (s *Service) generateAccountNumber(accountType AccountType) string {
	prefix := "62"
	switch accountType {
	case AccountTypeSavings:
		prefix += "10"
	case AccountTypeFixedDeposit:
		prefix += "20"
	case AccountTypeChecking:
		prefix += "30"
	}

	ts := time.Now().Format("20060102")
	suffix := fmt.Sprintf("%08d", time.Now().UnixNano()%100000000)
	return prefix + ts + suffix[:8]
}

func (s *Service) GetDB() *gorm.DB {
	return s.repo.db
}

func (r *Repository) DB() *gorm.DB {
	return r.db
}
