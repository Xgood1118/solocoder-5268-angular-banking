package transfer

import (
	"banking/internal/account"
	"banking/internal/audit"
	"banking/internal/limit"
	"errors"
	"fmt"
	"log"
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

	var toAccountID uint
	var toAccountNo string
	var toAccountName string

	if req.TransferType == TransferTypeIntraBank {
		if req.ToAccountID > 0 {
			toAcc, err := s.accountSvc.GetAccountByID(req.ToAccountID)
			if err != nil {
				return nil, errors.New("to account not found")
			}
			toAccountID = toAcc.ID
			toAccountNo = toAcc.AccountNumber
			toAccountName = toAcc.AccountName
		} else if req.ToAccountNo != "" {
			toAcc, err := s.accountSvc.GetAccountByNumber(req.ToAccountNo)
			if err != nil {
				return nil, errors.New("to account not found")
			}
			toAccountID = toAcc.ID
			toAccountNo = toAcc.AccountNumber
			toAccountName = toAcc.AccountName
		} else {
			return nil, errors.New("to_account_id or to_account_no is required for intra bank transfer")
		}

		if toAccountID == req.FromAccountID {
			return nil, errors.New("cannot transfer to the same account")
		}
	} else {
		toAccountNo = req.ToAccountNo
		toAccountName = req.ToAccountName
	}

	transfer := &Transfer{
		BizID:         req.BizID,
		UserID:        req.UserID,
		FromAccountID: req.FromAccountID,
		ToAccountID:   toAccountID,
		FromAccountNo: fromAcc.AccountNumber,
		ToAccountNo:   toAccountNo,
		ToBankName:    req.ToBankName,
		ToAccountName: toAccountName,
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

	return tx.Save(transfer).Error
}

func (s *Service) processInterBank(tx *gorm.DB, transfer *Transfer) error {
	if err := s.accountSvc.FreezeAmountTx(tx, transfer.FromAccountID, transfer.Amount+transfer.Fee); err != nil {
		return err
	}

	transfer.Status = StatusFrozen
	transfer.ClearingRefNo = uuid.New().String()

	if err := tx.Save(transfer).Error; err != nil {
		return err
	}

	go s.processInterBankAsync(transfer.ID)

	return nil
}

func (s *Service) processInterBankAsync(transferID uint) {
	time.Sleep(2 * time.Second)

	tx := s.repo.db.Begin()
	if tx.Error != nil {
		log.Printf("processInterBankAsync: begin tx failed: %v", tx.Error)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var t Transfer
	if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&t, transferID).Error; err != nil {
		tx.Rollback()
		log.Printf("processInterBankAsync: find transfer failed: %v", err)
		return
	}

	if t.Status != StatusFrozen {
		tx.Rollback()
		return
	}

	fmt.Printf("[MOCK] 发起跨行清算: %s, 金额: %.2f\n", t.ClearingRefNo, t.Amount)

	success := true
	if t.Amount > 50000 {
		success = false
	}

	if success {
		fmt.Printf("[MOCK] 跨行清算成功: %s\n", t.ClearingRefNo)

		if err := s.accountSvc.Debit(tx, t.FromAccountID, t.Amount,
			t.BizID+"-debit", t.Description, fmt.Sprintf("%d", t.ID)); err != nil {
			tx.Rollback()
			log.Printf("processInterBankAsync: debit failed: %v", err)
			return
		}

		if t.Fee > 0 {
			if err := s.accountSvc.Debit(tx, t.FromAccountID, t.Fee,
				t.BizID+"-fee", "手续费", fmt.Sprintf("%d", t.ID)); err != nil {
				tx.Rollback()
				log.Printf("processInterBankAsync: debit fee failed: %v", err)
				return
			}
		}

		if err := s.accountSvc.UnfreezeAmountTx(tx, t.FromAccountID, t.Amount+t.Fee); err != nil {
			tx.Rollback()
			log.Printf("processInterBankAsync: unfreeze failed: %v", err)
			return
		}

		now := time.Now()
		t.Status = StatusSuccess
		t.CompletedAt = &now
	} else {
		fmt.Printf("[MOCK] 跨行清算失败: %s\n", t.ClearingRefNo)

		if err := s.accountSvc.UnfreezeAmountTx(tx, t.FromAccountID, t.Amount+t.Fee); err != nil {
			tx.Rollback()
			log.Printf("processInterBankAsync: unfreeze failed: %v", err)
			return
		}

		t.Status = StatusFailed
		t.FailureReason = "清算失败，对方银行返回错误"
	}

	if err := tx.Save(&t).Error; err != nil {
		tx.Rollback()
		log.Printf("processInterBankAsync: save transfer failed: %v", err)
		return
	}

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
