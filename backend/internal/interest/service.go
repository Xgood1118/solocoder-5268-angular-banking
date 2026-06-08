package interest

import (
	"banking/internal/account"
	"log"
	"time"

	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

type Service struct {
	db         *gorm.DB
	accountSvc *account.Service
	cron       *cron.Cron
}

func NewService(db *gorm.DB, accountSvc *account.Service) *Service {
	return &Service{
		db:         db,
		accountSvc: accountSvc,
		cron:       cron.New(),
	}
}

func (s *Service) StartScheduler() {
	s.cron.AddFunc("0 0 0 * * *", func() {
		s.calculateDailyInterest()
	})
	s.cron.Start()
	log.Println("Interest scheduler started")
}

func (s *Service) calculateDailyInterest() {
	var accounts []account.Account
	s.accountSvc.GetDB().Where("status = ? AND interest_rate > 0", "active").Find(&accounts)

	for _, acc := range accounts {
		dailyRate := acc.InterestRate / 360

		dailyInterest := acc.Balance * dailyRate

		log.Printf("Account %d daily interest: %.4f", acc.ID, dailyInterest)
	}

	log.Printf("Daily interest calculated for %d accounts", len(accounts))
}

func (s *Service) calculateMonthlySettlement() {
	var accounts []account.Account
	s.accountSvc.GetDB().Where("status = ? AND account_type IN ?",
		"active", []string{"savings", "checking"}).Find(&accounts)

	tx := s.accountSvc.GetDB().Begin()

	for _, acc := range accounts {
		monthlyRate := acc.InterestRate / 12
		interest := acc.Balance * monthlyRate

		if interest < 0.01 {
			continue
		}

		bizID := "interest_" + time.Now().Format("200601") + "_" + string(rune(acc.ID))

		s.accountSvc.Credit(tx, acc.ID, interest, bizID, "利息结算", "interest", 0)
	}

	tx.Commit()
}

func (s *Service) CalculateInterest(accountID uint, days int) (float64, error) {
	acc, err := s.accountSvc.GetAccount(0, accountID)
	if err != nil {
		return 0, err
	}

	dailyRate := acc.InterestRate / 360
	interest := acc.Balance * dailyRate * float64(days)

	return interest, nil
}
