package auth

import (
	"banking/config"
	"banking/internal/audit"
	"banking/pkg/cache"
	"banking/pkg/masking"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
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
	repo    *Repository
	auditSvc *audit.Service
	pepper  string
	jwtKey  []byte
}

func NewService(repo *Repository, auditSvc *audit.Service) *Service {
	cfg := config.Load()
	return &Service{
		repo:    repo,
		auditSvc: auditSvc,
		pepper:  cfg.Pepper,
		jwtKey:  []byte(cfg.JWTSecret),
	}
}

func (s *Service) Register(req *RegisterRequest) (*User, error) {
	var count int64
	s.repo.db.Model(&User{}).Where("username = ?", req.Username).Count(&count)
	if count > 0 {
		return nil, errors.New("username already exists")
	}

	passwordHash, err := s.hashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user := &User{
		Username:     req.Username,
		PasswordHash: passwordHash,
		Email:        req.Email,
		Phone:        req.Phone,
		FullName:     req.FullName,
		IDCard:       req.IDCard,
		Status:       "active",
	}

	if err := s.repo.db.Create(user).Error; err != nil {
		return nil, err
	}

	s.auditSvc.Log(user.ID, "user_register", "auth", fmt.Sprintf("user %s registered", user.Username))

	return user, nil
}

func (s *Service) Login(req *LoginRequest) (*LoginResponse, error) {
	var user User
	if err := s.repo.db.Where("username = ?", req.Username).First(&user).Error; err != nil {
		return nil, errors.New("invalid username or password")
	}

	if user.Status != "active" {
		return nil, errors.New("account is locked")
	}

	if !s.checkPassword(req.Password, user.PasswordHash) {
		s.auditSvc.Log(user.ID, "login_failed", "auth", "invalid password")
		return nil, errors.New("invalid username or password")
	}

	if user.TwoFAEnabled {
		sendToken := uuid.New().String()
		s.repo.cache.Set(context.Background(), "twofa_pending:"+sendToken, fmt.Sprintf("%d", user.ID), 5*time.Minute)
		return &LoginResponse{
			NeedTwoFA:  true,
			TwoFAToken: sendToken,
		}, nil
	}

	token, err := s.generateToken(user.ID)
	if err != nil {
		return nil, err
	}

	s.auditSvc.Log(user.ID, "login_success", "auth", "user logged in")

	return &LoginResponse{
		Token: token,
		User:  &user,
	}, nil
}

func (s *Service) SendVerificationCode(userID uint, codeType, target string) error {
	code, err := s.generateCode()
	if err != nil {
		return err
	}

	lastCodeKey := fmt.Sprintf("last_code:%d:%s", userID, codeType)
	lastCode, _ := s.repo.cache.Get(context.Background(), lastCodeKey)
	if lastCode == code {
		code, _ = s.generateCode()
	}

	vc := &VerificationCode{
		UserID:    userID,
		Type:      codeType,
		Code:      code,
		Target:    target,
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}

	if err := s.repo.db.Create(vc).Error; err != nil {
		return err
	}

	s.repo.cache.Set(context.Background(), lastCodeKey, code, 5*time.Minute)

	fmt.Printf("[MOCK] Send %s code to %s: %s\n", codeType, target, code)

	s.auditSvc.Log(userID, "verification_code_sent", "auth", fmt.Sprintf("code sent to %s", masking.MaskPhone(target)))

	return nil
}

func (s *Service) VerifyCode(userID uint, codeType, code string) (bool, error) {
	var vc VerificationCode
	err := s.repo.db.Where("user_id = ? AND type = ? AND code = ? AND expires_at > ?",
		userID, codeType, code, time.Now()).First(&vc).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, errors.New("invalid or expired code")
		}
		return false, err
	}

	if vc.Locked {
		return false, errors.New("too many attempts, code locked")
	}

	if vc.Attempts >= 3 {
		s.repo.db.Model(&vc).Update("locked", true)
		return false, errors.New("too many attempts")
	}

	if vc.Code != code {
		s.repo.db.Model(&vc).UpdateColumn("attempts", vc.Attempts+1)
		return false, errors.New("invalid code")
	}

	s.repo.db.Delete(&vc)
	return true, nil
}

func (s *Service) VerifyTwoFA(twoFAToken, code string) (*LoginResponse, error) {
	userIDStr, err := s.repo.cache.Get(context.Background(), "twofa_pending:"+twoFAToken)
	if err != nil {
		return nil, errors.New("invalid or expired session")
	}

	var userID uint
	fmt.Sscanf(userIDStr, "%d", &userID)

	var user User
	if err := s.repo.db.First(&user, userID).Error; err != nil {
		return nil, errors.New("user not found")
	}

	valid, err := s.VerifyCode(userID, "sms", code)
	if err != nil || !valid {
		return nil, err
	}

	s.repo.cache.Del(context.Background(), "twofa_pending:"+twoFAToken)

	token, err := s.generateToken(userID)
	if err != nil {
		return nil, err
	}

	s.auditSvc.Log(userID, "twofa_success", "auth", "two factor authentication success")

	return &LoginResponse{
		Token: token,
		User:  &user,
	}, nil
}

func (s *Service) GetUser(id uint) (*User, error) {
	var user User
	if err := s.repo.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *Service) ChangePassword(userID uint, oldPassword, newPassword string) error {
	var user User
	if err := s.repo.db.First(&user, userID).Error; err != nil {
		return err
	}

	if !s.checkPassword(oldPassword, user.PasswordHash) {
		return errors.New("invalid old password")
	}

	newHash, err := s.hashPassword(newPassword)
	if err != nil {
		return err
	}

	user.PasswordHash = newHash
	s.repo.db.Save(&user)

	s.auditSvc.Log(userID, "password_changed", "auth", "password changed successfully")

	return nil
}

func (s *Service) hashPassword(password string) (string, error) {
	peppered := password + s.pepper
	hash, err := bcrypt.GenerateFromPassword([]byte(peppered), 12)
	return string(hash), err
}

func (s *Service) checkPassword(password, hash string) bool {
	peppered := password + s.pepper
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(peppered))
	return err == nil
}

func (s *Service) generateCode() (string, error) {
	code := ""
	for i := 0; i < 6; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		code += n.String()
	}
	return code, nil
}

func (s *Service) generateToken(userID uint) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtKey)
}

func (s *Service) ValidateToken(tokenStr string) (uint, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtKey, nil
	})

	if err != nil {
		return 0, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID := uint(claims["user_id"].(float64))
		return userID, nil
	}

	return 0, errors.New("invalid token")
}

func GenerateHMAC(data, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}
