package limit

import (
	"banking/pkg/cache"
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type Repository struct {
	db    *gorm.DB
	cache cache.Cache
}

func NewRepository(db *gorm.DB, c cache.Cache) *Repository {
	return &Repository{db: db, cache: c}
}

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CheckLimit(userID uint, amount float64, scope string) error {
	defaultPerTx := 50000.0
	defaultDaily := 200000.0
	defaultMonthly := 500000.0

	if amount > defaultPerTx {
		return fmt.Errorf("单笔限额 %.2f 元，当前金额 %.2f 元超限", defaultPerTx, amount)
	}

	dailyKey := fmt.Sprintf("limit:daily:%d:%s", userID, scope)
	dailyUsed, _ := s.repo.cache.GetFloat64(context.Background(), dailyKey)
	if dailyUsed+amount > defaultDaily {
		return fmt.Errorf("今日累计限额 %.2f 元，已用 %.2f 元，本次 %.2f 元超限", defaultDaily, dailyUsed, amount)
	}

	monthlyKey := fmt.Sprintf("limit:monthly:%d:%s", userID, scope)
	monthlyUsed, _ := s.repo.cache.GetFloat64(context.Background(), monthlyKey)
	if monthlyUsed+amount > defaultMonthly {
		return fmt.Errorf("本月累计限额 %.2f 元，已用 %.2f 元，本次 %.2f 元超限", defaultMonthly, monthlyUsed, amount)
	}

	return nil
}

func (s *Service) IncrementUsage(userID uint, amount float64, scope string) {
	dailyKey := fmt.Sprintf("limit:daily:%d:%s", userID, scope)
	ctx := context.Background()

	s.repo.cache.IncrByFloat(ctx, dailyKey, amount)
	s.repo.cache.Expire(ctx, dailyKey, 24*time.Hour)

	monthlyKey := fmt.Sprintf("limit:monthly:%d:%s", userID, scope)
	s.repo.cache.IncrByFloat(ctx, monthlyKey, amount)

	now := time.Now()
	nextMonth := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, now.Location())
	s.repo.cache.ExpireAt(ctx, monthlyKey, nextMonth)
}

func (s *Service) GetLimits(userID uint, scope string) map[string]interface{} {
	dailyKey := fmt.Sprintf("limit:daily:%d:%s", userID, scope)
	monthlyKey := fmt.Sprintf("limit:monthly:%d:%s", userID, scope)
	ctx := context.Background()

	dailyUsed, _ := s.repo.cache.GetFloat64(ctx, dailyKey)
	monthlyUsed, _ := s.repo.cache.GetFloat64(ctx, monthlyKey)

	return map[string]interface{}{
		"per_transaction": map[string]float64{
			"limit": 50000.0,
			"used":  0,
		},
		"daily": map[string]float64{
			"limit": 200000.0,
			"used":  dailyUsed,
		},
		"monthly": map[string]float64{
			"limit": 500000.0,
			"used":  monthlyUsed,
		},
	}
}

func (s *Service) SetLimit(userID uint, limitType LimitType, amount float64, scope string) error {
	if amount <= 0 {
		return errors.New("limit amount must be positive")
	}

	var config LimitConfig
	s.repo.db.Where("user_id = ? AND limit_type = ? AND scope = ?", userID, limitType, scope).First(&config)

	if config.ID == 0 {
		config = LimitConfig{
			UserID:    userID,
			LimitType: limitType,
			Amount:    amount,
			Scope:     scope,
		}
		return s.repo.db.Create(&config).Error
	}

	config.Amount = amount
	return s.repo.db.Save(&config).Error
}

func (s *Service) GetUserLimitConfigs(userID uint) ([]LimitConfig, error) {
	var configs []LimitConfig
	err := s.repo.db.Where("user_id = ? OR user_id = 0", userID).Find(&configs).Error
	return configs, err
}
