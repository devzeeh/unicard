package auth

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"math/big"
	"strings"
	"sync"
	"time"
	"unicard-go/backend/internal/pkg/account"
	"unicard-go/backend/internal/pkg/storage"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

// otpStore is an in-memory map of email OTPData.
// TODO: replace with Redis for multi-instance safety.
var (
	otpStore   = make(map[string]OTPData)
	otpStoreMu sync.Mutex
)

// Service holds all authentication business logic.
type Service struct {
	repo    *Repository
	storage storage.Service
}

func NewService(repo *Repository, storage storage.Service) *Service {
	return &Service{repo: repo, storage: storage}
}

// ---------------------------------------------------------------------------
// Login
// ---------------------------------------------------------------------------

type LoginResult struct {
	ID          string
	Username    string
	Role        string
	RedirectURL string
	Tokens      struct {
		Access  string
		Refresh string
	}
}

func (s *Service) Login(req LoginRequest) (LoginResult, error) {
	user, err := s.repo.FindUserByIdentifier(req.Identifier)
	if err != nil {
		log.Printf("login: user lookup failed for %q: %v", req.Identifier, err)
		return LoginResult{}, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		log.Printf("login: password mismatch for %q", req.Identifier)
		return LoginResult{}, ErrInvalidCredentials
	}

	access, refresh, err := GenerateTokens(user.UserID, user.Role)
	if err != nil {
		return LoginResult{}, fmt.Errorf("generate tokens: %w", err)
	}

	result := LoginResult{
		ID:          user.UserID,
		Username:    user.Username,
		Role:        user.Role,
		RedirectURL: redirectFor(user.Role, user.Username),
	}
	result.Tokens.Access = access
	result.Tokens.Refresh = refresh
	return result, nil
}

func redirectFor(role, username string) string {
	switch role {
	case "super_admin":
		return "/admin/" + username
	case "merchant_admin", "merchant_staff":
		return "/merchant/" + username + "/dashboard"
	default:
		return "/u/" + username + "/dashboard"
	}
}

// ---------------------------------------------------------------------------
// Customer signup
// ---------------------------------------------------------------------------

const (
	cardStatusInactive = "inactive"
	userRoleCustomer   = "customer"
	userStatusActive   = "active"
)

func (s *Service) SignupSendOTP(email, contactNumber string) error {
	exists, err := account.IsEmailExist(s.repo.store.DB(), email)
	if err != nil {
		return fmt.Errorf("email check: %w", err)
	}
	if exists {
		return errors.New("email already registered")
	}

	phoneExists, err := s.repo.IsPhoneExist(contactNumber)
	if err != nil {
		return fmt.Errorf("phone check: %w", err)
	}
	if phoneExists {
		return errors.New("phone number already registered")
	}

	otp := generateOTP()
	setOTP(email, otp)

	if err := SendSignupOTPEmail(email, "User", otp); err != nil {
		log.Printf("SignupSendOTP: failed to send OTP to %s: %v", email, err)
		return fmt.Errorf("send OTP email: %w", err)
	}
	return nil
}

func (s *Service) SignupVerifyOTP(email, otp string) error {
	return verifyOTP(email, otp)
}

func (s *Service) CheckCard(cardNumber string) error {
	status, err := s.repo.GetCardStatus(cardNumber)
	if err != nil {
		return errors.New("card not found")
	}
	if status != cardStatusInactive {
		return errors.New("card is invalid")
	}
	return nil
}

func (s *Service) Signup(ctx context.Context, req SignupRequest) error {
	hashedPassword, err := account.HashPassword(req.Password)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	userID, err := GenerateUserID(s.repo)
	if err != nil {
		return fmt.Errorf("generate user ID: %w", err)
	}

	createdAt, err := CurrentTimestamp()
	if err != nil {
		return fmt.Errorf("get timestamp: %w", err)
	}

	balance, err := s.repo.GetInitialBalance(req.CardNumber)
	if err != nil {
		return errors.New("invalid card number")
	}

	userIDStr := fmt.Sprintf("%d", userID)
	fullName := strings.TrimSpace(req.FirstName) + " " + strings.TrimSpace(req.LastName)

	err = s.repo.store.ExecTx(ctx, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO users (user_id, username, name, email, phone_number, password_hash, role, status, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			userIDStr, req.FirstName, fullName, req.Email, req.ContactNumber,
			hashedPassword, userRoleCustomer, userStatusActive, createdAt,
		)
		if err != nil {
			return fmt.Errorf("insert user: %w", err)
		}

		_, err = tx.ExecContext(ctx, `
			UPDATE cards
			SET status = 'active', user_id = ?, card_type = 'regular',
			    linked_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP,
			    expiry_date = DATE_ADD(CURRENT_DATE, INTERVAL 5 YEAR)
			WHERE card_number = ?`,
			userIDStr, req.CardNumber,
		)
		if err != nil {
			return fmt.Errorf("activate card: %w", err)
		}
		return nil
	})
	if err != nil {
		return err
	}

	log.Printf("Signup: account created — userID=%s balance=%.2f", userIDStr, balance)

	go func() {
		if err := SendWelcomeEmail(req.Email, fullName); err != nil {
			log.Printf("Signup: welcome email failed for %s: %v", req.Email, err)
		}
	}()

	return nil
}

// ---------------------------------------------------------------------------
// Admin signup
// ---------------------------------------------------------------------------

func (s *Service) AdminSignup(ctx context.Context, req AdminSignupRequest) error {
	req.Name = strings.TrimSpace(req.Name)
	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(req.Email)
	req.Password = strings.TrimSpace(req.Password)

	if req.Name == "" || req.Username == "" || req.Email == "" || req.Password == "" {
		return errors.New("all fields are required")
	}

	hashedPassword, err := account.HashPassword(req.Password)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	userID, err := GenerateUserID(s.repo)
	if err != nil {
		return fmt.Errorf("generate user ID: %w", err)
	}

	_, err = s.repo.store.ExecContext(ctx, `
		INSERT INTO users (user_id, username, name, email, password_hash, role, status)
		VALUES (?, ?, ?, ?, ?, 'super_admin', 'active')`,
		fmt.Sprintf("%d", userID), req.Username, req.Name, req.Email, hashedPassword,
	)
	if err != nil {
		log.Printf("AdminSignup: DB insert failed: %v", err)
		return errors.New("email or username already taken")
	}
	return nil
}

// ---------------------------------------------------------------------------
// Merchant signup
// ---------------------------------------------------------------------------

func (s *Service) MerchantSignup(ctx context.Context, req MerchantSignupRequest) error {
	// Sanitise
	caser := cases.Title(language.English)
	req.BusinessName = caser.String(strings.ToLower(strings.TrimSpace(req.BusinessName)))
	req.BusinessAddress = caser.String(strings.ToLower(strings.TrimSpace(req.BusinessAddress)))
	req.OwnerName = caser.String(strings.ToLower(strings.TrimSpace(req.OwnerName)))
	req.BusinessEmail = strings.ToLower(strings.TrimSpace(req.BusinessEmail))
	req.BusinessPhone = strings.TrimSpace(req.BusinessPhone)
	req.BusinessType = strings.TrimSpace(req.BusinessType)
	req.Password = strings.TrimSpace(req.Password)

	exists, err := account.IsEmailExist(s.repo.store.DB(), req.BusinessEmail)
	if err != nil {
		return fmt.Errorf("email check: %w", err)
	}
	if exists {
		return errors.New("email already registered")
	}

	hashedPassword, err := account.HashPassword(req.Password)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	// Generate IDs
	timestamp := time.Now().Format("01020605")
	nUser, _ := rand.Int(rand.Reader, big.NewInt(10000))
	userID := fmt.Sprintf("UNI-%s%04d", timestamp, nUser.Int64())

	nMerchant, _ := rand.Int(rand.Reader, big.NewInt(10000))
	merchantID := fmt.Sprintf("MCH-%s%04d", timestamp, nMerchant.Int64())

	nReg, _ := rand.Int(rand.Reader, big.NewInt(10000000000))
	regNum := fmt.Sprintf("UCBZ-%s-%010d", time.Now().Format("010205"), nReg.Int64())

	// Upload documents
	bizDocPath := s.uploadBase64(ctx, req.BusinessDocument)
	birPath := s.uploadBase64(ctx, req.BirDocument)
	otherPath := s.uploadBase64(ctx, req.OtherDocument)

	tx, err := s.repo.store.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
		INSERT INTO users (user_id, username, name, email, phone_number, password_hash, role, status)
		VALUES (?, ?, ?, ?, ?, ?, 'merchant_admin', 'active')`,
		userID, req.BusinessEmail, req.OwnerName,
		req.BusinessEmail, req.BusinessPhone, string(hashedPassword),
	)
	if err != nil {
		log.Printf("MerchantSignup: insert user failed: %v", err)
		return errors.New("failed to create user account — email or phone may already exist")
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO merchants (
			merchant_id, business_name, business_type, business_registration_number,
			business_address, user_id, owner_name, business_email, business_phone,
			commission_rate, status, business_document, bir_document, valid_id
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'pending approval', ?, ?, ?)`,
		merchantID, req.BusinessName, req.BusinessType, regNum, req.BusinessAddress,
		userID, req.OwnerName, req.BusinessEmail, req.BusinessPhone, 2.00,
		bizDocPath, birPath, otherPath,
	)
	if err != nil {
		log.Printf("MerchantSignup: insert merchant failed: %v", err)
		return errors.New("failed to create merchant profile")
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

// uploadBase64 is a best-effort helper; returns "" on failure so the signup
// can still proceed without a document upload.
func (s *Service) uploadBase64(ctx context.Context, b64data string) string {
	if b64data == "" || s.storage == nil {
		return ""
	}
	url, err := s.storage.UploadBase64(ctx, b64data)
	if err != nil {
		log.Printf("uploadBase64: R2 upload failed: %v", err)
		return ""
	}
	return url
}

// ---------------------------------------------------------------------------
// Forgot / reset password
// ---------------------------------------------------------------------------

func (s *Service) ForgotPasswordSendOTP(ctx context.Context, email string) error {
	exists, err := account.IsEmailExist(s.repo.store.DB(), email)
	if err != nil {
		return fmt.Errorf("email check: %w", err)
	}
	// Always return nil to prevent email enumeration; caller responds with
	// a generic "if found, OTP was sent" message even when exists==false.
	if !exists {
		return nil
	}

	name := s.repo.FindNameByEmail(email)
	otp := generateOTP()
	setOTP(email, otp)

	if err := SendOTPEmail(email, name, otp); err != nil {
		log.Printf("ForgotPasswordSendOTP: send failed for %s: %v", email, err)
		return fmt.Errorf("send OTP email: %w", err)
	}
	return nil
}

func (s *Service) ForgotPasswordVerifyOTP(email, otp string) error {
	return verifyOTP(email, otp)
}

func (s *Service) ResetPassword(ctx context.Context, email, otp, newPassword string) error {
	if err := verifyOTP(email, otp); err != nil {
		return err
	}

	if err := ValidatePassword(newPassword); err != nil {
		return err
	}

	hashedPassword, err := account.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	if err := s.repo.UpdatePassword(email, hashedPassword); err != nil {
		return fmt.Errorf("update password: %w", err)
	}

	// Fire-and-forget: log activity + send confirmation email.
	name, userID, err := s.repo.FindUserByEmail(email)
	if err == nil {
		s.repo.InsertActivityLog(userID, "password_reset", "in_app", "completed", "Password reset successfully")
		go func() {
			if err := SendPasswordChangedEmail(email, name); err != nil {
				log.Printf("ResetPassword: confirmation email failed for %s: %v", email, err)
			}
		}()
	}

	deleteOTP(email)
	return nil
}

// ---------------------------------------------------------------------------
// OTP helpers (internal)
// ---------------------------------------------------------------------------

func generateOTP() string {
	n, _ := rand.Int(rand.Reader, big.NewInt(1_000_000))
	return fmt.Sprintf("%06d", n.Int64())
}

func setOTP(email, otp string) {
	otpStoreMu.Lock()
	defer otpStoreMu.Unlock()
	otpStore[email] = OTPData{OTP: otp, Expiry: time.Now().Add(5 * time.Minute)}
}

func verifyOTP(email, otp string) error {
	otpStoreMu.Lock()
	defer otpStoreMu.Unlock()

	data, ok := otpStore[email]
	if !ok || data.OTP != otp {
		return errors.New("invalid OTP")
	}
	if time.Now().After(data.Expiry) {
		delete(otpStore, email)
		return errors.New("OTP expired")
	}
	return nil
}

func deleteOTP(email string) {
	otpStoreMu.Lock()
	defer otpStoreMu.Unlock()
	delete(otpStore, email)
}
