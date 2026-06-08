package audit

import (
	"banking/config"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

type Service struct {
	repo    *Repository
	hmacKey string
}

func NewService(repo *Repository) *Service {
	cfg := config.Load()
	return &Service{
		repo:    repo,
		hmacKey: cfg.AuditHMACKey,
	}
}

func (s *Service) Log(userID uint, action, module, description string) error {
	return s.LogWithIP(userID, action, module, description, "", "")
}

func (s *Service) LogWithIP(userID uint, action, module, description, ipAddress, userAgent string) error {
	var lastLog AuditLog
	s.repo.db.Order("id DESC").First(&lastLog)

	log := &AuditLog{
		UserID:      userID,
		Action:      action,
		Module:      module,
		Description: description,
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
	}

	data := fmt.Sprintf("%d|%s|%s|%s|%d|%s",
		userID, action, module, description, lastLog.ID, lastLog.HMAC)
	log.HMAC = s.generateHMAC(data)
	log.PrevLogHMAC = lastLog.HMAC

	return s.repo.db.Create(log).Error
}

func (s *Service) Verify(log *AuditLog, prevLog *AuditLog) bool {
	data := fmt.Sprintf("%d|%s|%s|%s|%d|%s",
		log.UserID, log.Action, log.Module, log.Description,
		func() uint {
			if prevLog != nil {
				return prevLog.ID
			}
			return 0
		}(),
		func() string {
			if prevLog != nil {
				return prevLog.HMAC
			}
			return ""
		}(),
	)

	calculated := s.generateHMAC(data)
	return hmac.Equal([]byte(calculated), []byte(log.HMAC))
}

func (s *Service) Query(query *AuditQuery) (*AuditLogResponse, error) {
	db := s.repo.db.Model(&AuditLog{})

	if query.UserID > 0 {
		db = db.Where("user_id = ?", query.UserID)
	}
	if query.Action != "" {
		db = db.Where("action = ?", query.Action)
	}
	if query.Module != "" {
		db = db.Where("module = ?", query.Module)
	}
	if !query.StartTime.IsZero() {
		db = db.Where("created_at >= ?", query.StartTime)
	}
	if !query.EndTime.IsZero() {
		db = db.Where("created_at <= ?", query.EndTime)
	}

	var total int64
	db.Count(&total)

	var logs []AuditLog
	page := query.Page
	if page <= 0 {
		page = 1
	}
	pageSize := query.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs)

	return &AuditLogResponse{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Logs:     logs,
	}, nil
}

func (s *Service) generateHMAC(data string) string {
	h := hmac.New(sha256.New, []byte(s.hmacKey))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

func (s *Service) CleanOldLogs(retentionDays int) error {
	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	return s.repo.db.Where("created_at < ?", cutoff).Delete(&AuditLog{}).Error
}
