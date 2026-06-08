package transfer

import (
	"banking/internal/account"
	"banking/internal/audit"
	"banking/internal/limit"
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
	repo       *Repository
	accountSvc *account.Service
	auditSvc   *audit.Service
	limitSvc   *limit.Service
}

func NewService(repo *Repository, accountSvc *account.Service, auditSvc *audit.Service, limitSvc *limit.Service) *Service {
	return &Service{
		repo:       repo,
		accountSvc: accountSvc,
		auditSvc:   auditSvc,
		limitSvc:   limitSvc,
	}
}

func (s *Service) CreateTransfer(req *CreateTransferRequest) (*Transfer, error) {
	var existing Transfer
	err := s.repo.db.Where("biz_id = ?", req.BizID).First(&existing).Error
	if err == nil {
		return &existing, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if req.Amount <= 0 {
		return nil, errors.New("amount must be greater than 0")
	}

	fromAcc, err := s.accountSvc.GetAccount(req.UserID, req.FromAccountID)
	if err != nil {
		return nil, errors.New("from account not found")
	}

	if fromAcc.Status != "active" {
		return nil, errors.New("from account is not active")
	}

	if err := s.limitSvc.CheckLimit(req.UserID, req.Amount, "transfer"); err != nil {
		return nil, err
	}

	fee := s.calculateFee(req.Amount, req.TransferType, req.TransferSpeed)
	totalAmount := req.Amount + fee

	if fromAcc.AvailableBalance < totalAmount {
		return nil, errors.New("insufficient balance")
	}

	transfer := &Transfer{
		BizID:         req.BizID,
		UserID:        req.UserID,
		FromAccountID: req.FromAccountID,
		ToAccountID:   req.ToAccountID,
		FromAccountNo: fromAcc.AccountNumber,
		ToAccountNo:   req.ToAccountNo,
		ToBankName:    req.ToBankName,
		ToAccountName: req.ToAccountName,
		Amount:        req.Amount,
		Currency:      string(fromAcc.Currency),
		TransferType:  req.TransferType,
		TransferSpeed: req.TransferSpeed,
		Status:        StatusPending,
		Description:   req.Description,
		Fee:           fee,
	}

	tx := s.repo.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	if err := tx.Create(transfer).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if req.TransferType == TransferTypeIntraBank {
		err = s.processIntraBank(tx, transfer)
	} else {
		err = s.processInterBank(tx, transfer)
	}

	if err != nil {
		transfer.Status = StatusFailed
		transfer.FailureReason = err.Error()
		tx.Save(transfer)
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	s.limitSvc.IncrementUsage(req.UserID, req.Amount, "transfer")

	s.auditSvc.Log(req.UserID, "transfer", "transfer",
		fmt.Sprintf("transfer %s: %.2f %s -> %s", transfer.BizID, transfer.Amount, fromAcc.AccountNumber, req.ToAccountName))

	return transfer, nil
}

func (s *Service) processIntraBank(tx *gorm.DB, transfer *Transfer) error {
	if transfer.ToAccountID == 0 {
		return errors.New("to account id required for intra bank transfer")
	}

	toAcc, err := s.accountSvc.GetAccount(0, transfer.ToAccountID)
	if err != nil {
		return errors.New("to account not found")
	}
	_ = toAcc

	if err := s.accountSvc.Debit(tx, transfer.FromAccountID, transfer.Amount,
		transfer.BizID+"-debit", transfer.Description, fmt.Sprintf("%d", transfer.ID)); err != nil {
		return err
	}

	if transfer.Fee > 0 {
		if err := s.accountSvc.Debit(tx, transfer.FromAccountID, transfer.Fee,
			transfer.BizID+"-fee", "手续费", fmt.Sprintf("%d", transfer.ID)); err != nil {
			return err
		}
	}

	if err := s.accountSvc.Credit(tx, transfer.ToAccountID, transfer.Amount,
		transfer.BizID+"-credit", transfer.Description, fmt.Sprintf("%d", transfer.ID), transfer.FromAccountID); err != nil {
		return err
	}

	now := time.Now()
	transfer.Status = StatusSuccess
	transfer.CompletedAt = &now

	return nil
}

func (s *Service) processInterBank(tx *gorm.DB, transfer *Transfer) error {
	if err := s.accountSvc.FreezeAmount(transfer.FromAccountID, transfer.Amount+transfer.Fee, transfer.BizID); err != nil {
		return err
	}

	transfer.Status = StatusFrozen
	transfer.ClearingRefNo = uuid.New().String()

	go s.processInterBankAsync(transfer)

	return nil
}

func (s *Service) processInterBankAsync(transfer *Transfer) {
	time.Sleep(2 * time.Second)

	fmt.Printf("[MOCK] 发起跨行清算: %s, 金额: %.2f\n", transfer.ClearingRefNo, transfer.Amount)

	success := true
	// 模拟90%成功率
	if transfer.Amount > 50000 {
		success = false
	}

	tx := s.repo.db.Begin()

	var t Transfer
	tx.Where("id = ?", transfer.ID).First(&t)

	if success {
		fmt.Printf("[MOCK] 跨行清算成功: %s\n", transfer.ClearingRefNo)

		if err := s.accountSvc.Debit(tx, t.FromAccountID, t.Amount,
			t.BizID+"-debit", t.Description, fmt.Sprintf("%d", t.ID)); err != nil {
			tx.Rollback()
			return
		}

		if t.Fee > 0 {
			if err := s.accountSvc.Debit(tx, t.FromAccountID, t.Fee,
				t.BizID+"-fee", "手续费", fmt.Sprintf("%d", t.ID)); err != nil {
				tx.Rollback()
				return
			}
		}

		if err := s.accountSvc.UnfreezeAmount(t.FromAccountID, t.Amount+t.Fee); err != nil {
			tx.Rollback()
			return
		}

		now := time.Now()
		t.Status = StatusSuccess
		t.CompletedAt = &now
	} else {
		fmt.Printf("[MOCK] 跨行清算失败: %s\n", transfer.ClearingRefNo)

		if err := s.accountSvc.UnfreezeAmount(t.FromAccountID, t.Amount+t.Fee); err != nil {
			tx.Rollback()
			return
		}

		t.Status = StatusFailed
		t.FailureReason = "清算失败，对方银行返回错误"
	}

	tx.Save(&t)
	tx.Commit()
}

func (s *Service) calculateFee(amount float64, transferType TransferType, speed TransferSpeed) float64 {
	feeRate := 0.001
	if transferType == TransferTypeInterBank {
		feeRate = 0.002
	}
	if speed == TransferSpeedRealTime {
		feeRate += 0.001
	}

	fee := amount * feeRate

	switch transferType {
	case TransferTypeIntraBank:
		if fee < 0 {
			fee = 0
		}
		if fee > 50 {
			fee = 50
		}
	case TransferTypeInterBank:
		if fee < 2 {
			fee = 2
		}
		if fee > 50 {
			fee = 50
		}
	}

	return fee
}

func (s *Service) GetTransfer(userID uint, id uint) (*Transfer, error) {
	var transfer Transfer
	if err := s.repo.db.First(&transfer, id).Error; err != nil {
		return nil, err
	}

	if transfer.UserID != userID {
		return nil, errors.New("permission denied")
	}

	return &transfer, nil
}

func (s *Service) ListTransfers(userID uint, page, pageSize int, status string) (int64, []Transfer, error) {
	var total int64
	var transfers []Transfer

	db := s.repo.db.Model(&Transfer{}).Where("user_id = ?", userID)
	if status != "" {
		db = db.Where("status = ?", status)
	}

	db.Count(&total)

	offset := (page - 1) * pageSize
	err := db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&transfers).Error

	return total, transfers, err
}

func (s *Service) GetByBizID(userID uint, bizID string) (*Transfer, error) {
	var transfer Transfer
	if err := s.repo.db.Where("biz_id = ?", bizID).First(&transfer).Error; err != nil {
		return nil, err
	}

	if transfer.UserID != userID {
		return nil, errors.New("permission denied")
	}

	return &transfer, nil
}
